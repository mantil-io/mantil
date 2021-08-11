package deploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/commands"
	"github.com/atoz-technology/mantil-cli/internal/docker"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/atoz-technology/mantil-cli/internal/shell"
	"github.com/atoz-technology/mantil-cli/internal/util"
)

const (
	FunctionsDir = "functions"
	BuildDir     = "build"
	BinaryName   = "bootstrap"
)

type DeployCmd struct {
	aws     *aws.AWS
	project *mantil.Project
	path    string
	token   string
}

func New(project *mantil.Project, awsClient *aws.AWS, path, token string) (*DeployCmd, error) {
	d := &DeployCmd{
		aws:     awsClient,
		project: project,
		path:    path,
		token:   token,
	}
	return d, nil
}

func (d *DeployCmd) Deploy() error {
	functionUpdates, err := d.deploySync()
	if err != nil {
		return err
	}
	if len(functionUpdates) == 0 {
		log.Printf("no function changes - nothing to deploy")
		return nil
	}
	if err = d.deployRequest(functionUpdates); err != nil {
		return err
	}
	return nil
}

func (d *DeployCmd) deploySync() ([]mantil.FunctionUpdate, error) {
	localFuncs, err := d.localFunctions()
	if err != nil {
		return nil, err
	}

	addedFuncs := d.processAddedFunctions(localFuncs)
	removedFuncs := d.processRemovedFunctions(localFuncs)
	funcsForDeploy := d.prepareFunctionsForDeploy()

	var functionUpdates []mantil.FunctionUpdate
	for _, f := range funcsForDeploy {
		added := false
		for _, af := range addedFuncs {
			if af == f.Name {
				added = true
				break
			}
		}
		fu := mantil.FunctionUpdate{
			Name:     f.Name,
			Hash:     f.Hash,
			S3Key:    f.S3Key,
			ImageKey: f.ImageKey,
			Added:    added,
			Updated:  !added,
		}
		functionUpdates = append(functionUpdates, fu)
	}

	for _, f := range removedFuncs {
		functionUpdates = append(functionUpdates, mantil.FunctionUpdate{
			Name:    f,
			Removed: true,
		})
	}

	return functionUpdates, nil
}

func (d *DeployCmd) localFunctions() ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(d.path, FunctionsDir))
	if err != nil {
		return nil, err
	}
	functions := []string{}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		functions = append(functions, f.Name())
	}
	return functions, nil
}

func (d *DeployCmd) processAddedFunctions(localFuncs []string) []string {
	addedFunctions := d.addedFunctions(localFuncs)
	if len(addedFunctions) > 0 {
		log.Printf("added - %s", strings.Join(addedFunctions, ","))
	}
	return addedFunctions
}

// compares local functions with the ones in project config and returns all that are newly added
func (d *DeployCmd) addedFunctions(localFuncs []string) []string {
	funcExistsInProject := func(name string) bool {
		for _, f := range d.project.Functions {
			if name == f.Name {
				return true
			}
		}
		return false
	}
	added := []string{}
	for _, fun := range localFuncs {
		if !funcExistsInProject(fun) {
			added = append(added, fun)
		}
	}
	return added
}

func (d *DeployCmd) processRemovedFunctions(localFuncs []string) []string {
	removedFunctions := d.removedFunctions(localFuncs)
	if len(removedFunctions) > 0 {
		log.Printf("removed - %s", strings.Join(removedFunctions, ","))
	}
	return removedFunctions
}

// compares local functions with the ones in project config and returns all that are newly removed
func (d *DeployCmd) removedFunctions(localFuncs []string) []string {
	funcExistsLocally := func(name string) bool {
		for _, f := range localFuncs {
			if name == f {
				return true
			}
		}
		return false
	}
	removed := []string{}
	for _, fun := range d.project.Functions {
		if !funcExistsLocally(fun.Name) {
			removed = append(removed, fun.Name)
		}
	}
	return removed
}

// prepareFunctionsForDeploy goes through project functions, checks which ones have changed
// and uploads new version to s3/image to ECR if necessary
func (d *DeployCmd) prepareFunctionsForDeploy() []mantil.Function {
	funcsForDeploy := []mantil.Function{}
	for i, f := range d.project.Functions {
		funcDir := path.Join(d.path, FunctionsDir, f.Name)
		isImage := d.isFunctionImage(funcDir)

		if err := d.buildFunction(BinaryName, funcDir); err != nil {
			log.Printf("skipping function %s due to error while building - %v", f.Name, err)
			continue
		}
		binaryPath := path.Join(funcDir, BinaryName)
		hash, err := util.FileHash(binaryPath)
		if err != nil {
			log.Printf("skipping function %s due to error while calculating binary hash - %v", f.Name, err)
			continue
		}

		if hash != f.Hash {
			f.Hash = hash

			if isImage {
				log.Printf("Dockerfile found - creating function %s as image package type", f.Name)
				image, err := docker.ProcessFunctionImage(d.aws, f, mantil.ProjectIdentifier(d.project.Name), funcDir)
				if err != nil {
					log.Printf("skipping function %s due to error while processing docker image - %v", f.Name, err)
					continue
				}
				f.SetImageKey(image)
			} else {
				log.Printf("creating function %s as zip package type", f.Name)
				f.SetS3Key(fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash))
				if err := d.uploadBinaryToS3(f.S3Key, binaryPath); err != nil {
					log.Printf("skipping function %s due to error while processing s3 file", err)
					continue
				}
			}

			d.project.Functions[i] = f
			funcsForDeploy = append(funcsForDeploy, f)
		}
	}
	return funcsForDeploy
}

func (d *DeployCmd) buildFunction(name, funcDir string) error {
	return shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name}, funcDir)
}

func (d *DeployCmd) isFunctionImage(funcDir string) bool {
	_, err := os.Stat(path.Join(funcDir, "Dockerfile"))
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Printf("could not detect if Dockerfile exists - processing function as zip package type")
		return false
	}
	return true
}

func (d *DeployCmd) uploadBinaryToS3(key, binaryPath string) error {
	buf, err := util.CreateZipForFile(binaryPath, BinaryName)
	if err != nil {
		return err
	}
	if err := d.aws.PutObjectToS3Bucket(d.project.Bucket, key, bytes.NewReader(buf)); err != nil {
		return err
	}
	return nil
}

func (d *DeployCmd) deployRequest(updates []mantil.FunctionUpdate) error {
	type req struct {
		ProjectName     string
		Token           string
		FunctionUpdates []mantil.FunctionUpdate
	}
	r := &req{
		ProjectName:     d.project.Name,
		Token:           d.token,
		FunctionUpdates: updates,
	}
	return commands.BackendRequest("deploy", r, nil)
}
