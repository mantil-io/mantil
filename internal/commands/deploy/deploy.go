package deploy

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/docker"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/atoz-technology/mantil-cli/internal/shell"
	"github.com/atoz-technology/mantil-cli/internal/util"
)

const (
	FunctionsDir = "functions"
	BuildDir     = "build"
)

type DeployCmd struct {
	aws     *aws.AWS
	project *mantil.Project
	path    string
}

func New(project *mantil.Project, path string) (*DeployCmd, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	return &DeployCmd{
		aws:     awsClient,
		project: project,
		path:    path,
	}, nil
}

func (d *DeployCmd) Deploy() error {
	functionUpdates, err := d.deploySync()
	if err != nil {
		return err
	}
	if err = d.deployRequest(functionUpdates); err != nil {
		return err
	}
	return nil
}

// build function into binary with the function's name
func (d *DeployCmd) buildFunction(name, funcDir string) error {
	return shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name}, funcDir)
}

func (d *DeployCmd) deploySync() ([]mantil.FunctionUpdate, error) {
	funcs, err := d.localFunctions()
	if err != nil {
		return nil, err
	}

	addedFunctions := d.addedFunctions(funcs)
	if len(addedFunctions) > 0 {
		log.Printf("added functions - %s", strings.Join(addedFunctions, ","))
	}
	for _, af := range addedFunctions {
		d.project.AddFunction(mantil.Function{Name: af})
	}

	removedFunctions := d.removedFunctions(funcs)
	if len(removedFunctions) > 0 {
		log.Printf("removed functions - %s", strings.Join(removedFunctions, ","))
	}
	for _, rf := range removedFunctions {
		d.project.RemoveFunction(rf)
	}

	funcsForDeploy := d.prepareFunctionsForDeploy()

	var functionUpdates []mantil.FunctionUpdate
	for _, f := range funcsForDeploy {
		added := false
		for _, af := range addedFunctions {
			if af == f.Name {
				added = true
				break
			}
		}
		removed := false
		for _, rf := range removedFunctions {
			if rf == f.Name {
				removed = true
				break
			}
		}
		fu := mantil.FunctionUpdate{
			Name:     f.Name,
			Hash:     f.Hash,
			S3Key:    f.S3Key,
			ImageKey: f.ImageKey,
			Added:    added,
			Removed:  removed,
			Updated:  !added && !removed,
		}
		functionUpdates = append(functionUpdates, fu)
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

// compares local functions with the ones in project config and returns all that are newly added
func (d *DeployCmd) addedFunctions(functions []string) []string {
	functionExists := func(name string) bool {
		for _, f := range d.project.Functions {
			if name == f.Name {
				return true
			}
		}
		return false
	}
	added := []string{}
	for _, fun := range functions {
		if !functionExists(fun) {
			added = append(added, fun)
		}
	}
	return added
}

// compares local functions with the ones in project config and returns all that are newly removed
func (d *DeployCmd) removedFunctions(functions []string) []string {
	functionExists := func(name string) bool {
		for _, f := range functions {
			if name == f {
				return true
			}
		}
		return false
	}
	removed := []string{}
	for _, fun := range d.project.Functions {
		if !functionExists(fun.Name) {
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

		if err := d.buildFunction(f.Name, funcDir); err != nil {
			log.Printf("skipping function %s due to error while building - %v", f.Name, err)
			continue
		}

		isImage := true
		if _, err := os.Stat(path.Join(funcDir, "Dockerfile")); os.IsNotExist(err) {
			isImage = false
		}

		binaryPath := path.Join(funcDir, f.Name)
		hash, err := util.FileHash(binaryPath)
		if err != nil {
			log.Printf("skipping function %s due to error while calculating binary hash - %v", f.Name, err)
			continue
		}

		if hash != f.Hash {
			f.Hash = hash

			if isImage {
				log.Printf("Dockerfile found - creating function %s as image", f.Name)
				image, err := docker.ProcessFunctionImage(f, mantil.ProjectIdentifier(d.project.Name), funcDir)
				if err != nil {
					log.Printf("skipping function %s due to error while processing docker image - %v", f.Name, err)
					continue
				}
				f.SetImageKey(image)
			} else {
				log.Printf("creating function %s from s3", f.Name)
				f.SetS3Key(fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash))
				if err := d.processFunctionS3(f, binaryPath); err != nil {
					log.Printf("skipping function %s due to error while processing s3 file", err)
				}
			}

			d.project.Functions[i] = f
			funcsForDeploy = append(funcsForDeploy, f)
		}
	}
	return funcsForDeploy
}

func (d *DeployCmd) processFunctionS3(f mantil.Function, binaryPath string) error {
	buf, err := util.CreateZipForFile(binaryPath, f.Name)
	if err != nil {
		return err
	}
	if err := d.aws.PutObjectToS3Bucket(d.project.Bucket, f.S3Key, bytes.NewReader(buf)); err != nil {
		return err
	}
	return nil
}

func (d *DeployCmd) deployRequest(updates []mantil.FunctionUpdate) error {
	type req struct {
		ProjectBucket   string
		FunctionUpdates []mantil.FunctionUpdate
	}
	url := "https://try.mantil.team/mantil-backend/deploy"
	r := &req{
		ProjectBucket:   d.project.Bucket,
		FunctionUpdates: updates,
	}
	buf, err := json.Marshal(r)
	if err != nil {
		return err
	}
	rsp, err := http.Post(url, "application/json", bytes.NewBuffer(buf))
	if err != nil {
		return err
	}
	fmt.Println(rsp.Body)
	return nil
}
