package deploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/mantil"
	"github.com/atoz-technology/mantil-cli/internal/shell"
	"github.com/atoz-technology/mantil-cli/internal/terraform"
	"github.com/atoz-technology/mantil-cli/internal/util"
)

const (
	FunctionsDir = "functions"
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
	funcsForDeploy, changesInInfrastructures, err := d.deploySync()
	if err != nil {
		return err
	}

	// if there are changes in infrastructure let terraform update all the necessary functions among other changes
	if changesInInfrastructures {
		log.Printf("applying terraform due to infrastructure changes")
		if err := d.applyInfrastructure(); err != nil {
			return err
		}
		return mantil.SaveProject(d.project)
	}

	// otherwise just update lambda functions directly
	for _, f := range funcsForDeploy {
		log.Printf("updating function %s", f.Name)
		if err := d.updateLambdaFunction(f); err != nil {
			log.Print(err)
		}
	}
	return mantil.SaveProject(d.project)
}

func (d *DeployCmd) applyInfrastructure() error {
	tf := terraform.New(d.path)
	if err := tf.ApplyForProject(d.project, false); err != nil {
		return fmt.Errorf("error while applying terraform for project %s - %v", d.project.Name, err)
	}
	return nil
}

func (d *DeployCmd) updateLambdaFunction(f mantil.Function) error {
	lambdaName := fmt.Sprintf("%s-mantil-team-%s-%s", d.project.Organization.Name, d.project.Name, f.Name)
	if err := d.aws.UpdateLambdaFunctionCodeFromS3(lambdaName, d.project.Bucket, f.S3Key); err != nil {
		return fmt.Errorf("could not update lambda %s due to error %v", lambdaName, err)
	}
	return nil
}

// build function into binary with the function's name
func (d *DeployCmd) buildFunction(name, funcDir string) error {
	return shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name}, funcDir)
}

func (d *DeployCmd) deploySync() ([]mantil.Function, bool, error) {
	funcs, err := d.localFunctions()
	if err != nil {
		return nil, false, err
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
	d.project.AddFunctionDefaults()

	changesInInfrastructure := len(addedFunctions) != 0 || len(removedFunctions) != 0
	return funcsForDeploy, changesInInfrastructure, nil
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
// and uploads new version to s3 if necessary
func (d *DeployCmd) prepareFunctionsForDeploy() []mantil.Function {
	funcsForDeploy := []mantil.Function{}
	for i, f := range d.project.Functions {
		funcBuildDir := path.Join(d.path, FunctionsDir, "build", f.Name)
		if err := d.buildFunction(f.Name, funcBuildDir); err != nil {
			log.Printf("skipping function %s due to error while building - %v", f.Name, err)
			continue
		}

		binaryPath := path.Join(funcBuildDir, f.Name)
		hash, err := util.FileHash(binaryPath)
		if err != nil {
			log.Printf("skipping function %s due to error while calculating binary hash - %v", f.Name, err)
			continue
		}

		if hash != f.Hash {
			f.Hash = hash
			f.S3Key = fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash)

			buf, err := util.CreateZipForFile(binaryPath, f.Name)
			if err != nil {
				log.Printf("skipping function %s due to error while zipping binary - %v", f.Name, err)
				continue
			}

			if err := d.aws.PutObjectToS3Bucket(d.project.Bucket, f.S3Key, bytes.NewReader(buf)); err != nil {
				log.Printf("skipping function %s due to error while uploading to s3 - %v", f.Name, err)
				continue
			}

			d.project.Functions[i] = f
			funcsForDeploy = append(funcsForDeploy, f)
		}
	}
	return funcsForDeploy
}
