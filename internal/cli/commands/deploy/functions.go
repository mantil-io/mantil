package deploy

import (
	"fmt"
	"path"
	"strings"

	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/mantil-io/mantil/internal/shell"
	"github.com/mantil-io/mantil/internal/util"
)

func (d *DeployCmd) functionUpdates() ([]mantil.ProjectUpdate, error) {
	localFuncs, err := d.localDirs(FunctionsDir)
	if err != nil {
		return nil, err
	}

	addedFuncs := d.processAddedFunctions(localFuncs)
	removedFuncs := d.processRemovedFunctions(localFuncs)
	funcsForDeploy := d.prepareFunctionsForDeploy()

	var updates []mantil.ProjectUpdate
	for _, f := range funcsForDeploy {
		added := false
		for _, af := range addedFuncs {
			if af == f.Name {
				added = true
				break
			}
		}
		var action mantil.UpdateAction
		if added {
			action = mantil.Add
		} else {
			action = mantil.Update
		}
		u := mantil.ProjectUpdate{
			Function: &mantil.FunctionUpdate{
				Name:  f.Name,
				Hash:  f.Hash,
				S3Key: f.S3Key,
			},
			Action: action,
		}
		updates = append(updates, u)
	}

	for _, f := range removedFuncs {
		updates = append(updates, mantil.ProjectUpdate{
			Function: &mantil.FunctionUpdate{
				Name: f,
			},
			Action: mantil.Remove,
		})
	}
	return updates, nil
}

func (d *DeployCmd) processAddedFunctions(localFuncs []string) []string {
	addedFunctions := d.addedFunctions(localFuncs)
	if len(addedFunctions) > 0 {
		log.Debug("added functions: %s", strings.Join(addedFunctions, ","))
		for _, af := range addedFunctions {
			d.project.AddFunction(mantil.Function{Name: af})
		}
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
		log.Debug("removed functions: %s", strings.Join(removedFunctions, ","))
		for _, rf := range removedFunctions {
			d.project.RemoveFunction(rf)
		}
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
// and uploads new version to s3 if necessary
func (d *DeployCmd) prepareFunctionsForDeploy() []mantil.Function {
	funcsForDeploy := []mantil.Function{}
	for i, f := range d.project.Functions {
		log.Info("building function %s", f.Name)
		funcDir := path.Join(d.path, FunctionsDir, f.Name)

		if err := d.buildFunction(BinaryName, funcDir); err != nil {
			log.Errorf("skipping function %s due to error while building - %v", f.Name, err)
			continue
		}
		binaryPath := path.Join(funcDir, BinaryName)
		hash, err := util.FileHash(binaryPath)
		if err != nil {
			log.Errorf("skipping function %s due to error while calculating binary hash - %v", f.Name, err)
			continue
		}

		if hash != f.Hash {
			f.Hash = hash

			log.Debug("creating function %s as zip package type", f.Name)
			f.SetS3Key(fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash))
			log.Debug("uploading function %s to s3", f.Name)
			if err := d.uploadBinaryToS3(f.S3Key, binaryPath); err != nil {
				log.Errorf("skipping function %s due to error while processing s3 file - %v", f.Name, err)
				continue
			}
			d.project.Functions[i] = f
			funcsForDeploy = append(funcsForDeploy, f)
		}
	}
	return funcsForDeploy
}

func (d *DeployCmd) buildFunction(name, funcDir string) error {
	return shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name, "--tags", "lambda.norpc"}, funcDir)
}

func (d *DeployCmd) uploadBinaryToS3(key, binaryPath string) error {
	buf, err := util.CreateZipForFile(binaryPath, BinaryName)
	if err != nil {
		return err
	}
	if err := d.aws.PutObjectToS3Bucket(d.project.Bucket, key, buf); err != nil {
		return err
	}
	return nil
}
