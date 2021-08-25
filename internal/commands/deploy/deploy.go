package deploy

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mantil-io/mantil-cli/internal/aws"
	"github.com/mantil-io/mantil-cli/internal/commands"
	"github.com/mantil-io/mantil-cli/internal/docker"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/mantil-io/mantil-cli/internal/mantil"
	"github.com/mantil-io/mantil-cli/internal/shell"
	"github.com/mantil-io/mantil-cli/internal/util"
)

const (
	FunctionsDir   = "functions"
	StaticSitesDir = "static"
	BuildDir       = "build"
	BinaryName     = "bootstrap"
)

type DeployCmd struct {
	aws     *aws.AWS
	project *mantil.Project
	config  *mantil.LocalProjectConfig
	path    string
	token   string
	updates []mantil.ProjectUpdate
}

func New(project *mantil.Project, config *mantil.LocalProjectConfig, awsClient *aws.AWS, path, token string) (*DeployCmd, error) {
	d := &DeployCmd{
		aws:     awsClient,
		project: project,
		config:  config,
		path:    path,
		token:   token,
	}
	return d, nil
}

func (d *DeployCmd) Deploy() error {
	if err := d.deploySync(); err != nil {
		return err
	}
	if !d.HasUpdates() {
		log.Info("no changes - nothing to deploy")
		return nil
	}
	p, err := d.deployRequest()
	if err != nil {
		return err
	}
	log.Notice("deploy successfully finished")
	if p.ApiURL != d.config.ApiURL {
		d.config.ApiURL = p.ApiURL
		return d.config.Save(d.path)
	}
	for _, s := range p.StaticWebsites {
		d.uploadStaticWebsite(s.Name, s.Bucket)
	}
	return nil
}

func (d *DeployCmd) HasUpdates() bool {
	return len(d.updates) > 0
}

func (d *DeployCmd) deploySync() error {
	var updates []mantil.ProjectUpdate
	fu, err := d.functionUpdates()
	if err != nil {
		return err
	}
	updates = append(updates, fu...)
	su, err := d.staticSiteUpdates()
	if err != nil {
		return err
	}
	updates = append(updates, su...)
	d.updates = updates
	return nil
}

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
				Name:     f.Name,
				Hash:     f.Hash,
				S3Key:    f.S3Key,
				ImageKey: f.ImageKey,
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
		log.Debug("added - %s", strings.Join(addedFunctions, ","))
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
		log.Debug("removed - %s", strings.Join(removedFunctions, ","))
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
// and uploads new version to s3/image to ECR if necessary
func (d *DeployCmd) prepareFunctionsForDeploy() []mantil.Function {
	funcsForDeploy := []mantil.Function{}
	for i, f := range d.project.Functions {
		log.Info("building function %s", f.Name)
		funcDir := path.Join(d.path, FunctionsDir, f.Name)
		isImage := d.isFunctionImage(funcDir)

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

			if isImage {
				log.Debug("Dockerfile found - creating function %s as image package type", f.Name)
				image, err := docker.ProcessFunctionImage(d.aws, f, mantil.ProjectIdentifier(d.project.Name), funcDir)
				if err != nil {
					log.Errorf("skipping function %s due to error while processing docker image - %v", f.Name, err)
					continue
				}
				f.SetImageKey(image)
			} else {
				log.Debug("creating function %s as zip package type", f.Name)
				f.SetS3Key(fmt.Sprintf("functions/%s-%s.zip", f.Name, f.Hash))
				log.Debug("uploading function %s to s3", f.Name)
				if err := d.uploadBinaryToS3(f.S3Key, binaryPath); err != nil {
					log.Errorf("skipping function %s due to error while processing s3 file - %v", f.Name, err)
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
	return shell.Exec([]string{"env", "GOOS=linux", "GOARCH=amd64", "go", "build", "-o", name, "--tags", "lambda.norpc"}, funcDir)
}

func (d *DeployCmd) isFunctionImage(funcDir string) bool {
	_, err := os.Stat(path.Join(funcDir, "Dockerfile"))
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		log.Debug("could not detect if Dockerfile exists - processing function as zip package type")
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

func (d *DeployCmd) staticSiteUpdates() ([]mantil.ProjectUpdate, error) {
	var updates []mantil.ProjectUpdate
	localSites, err := d.localDirs(StaticSitesDir)
	if err != nil {
		return nil, err
	}
	var added, removed []string
	for _, ls := range localSites {
		isAdded := true
		for _, s := range d.project.StaticWebsites {
			if ls == s.Name {
				isAdded = false
			}
		}
		if isAdded {
			added = append(added, ls)
		}
	}
	for _, s := range d.project.StaticWebsites {
		isRemoved := true
		for _, ls := range localSites {
			if ls == s.Name {
				isRemoved = false
			}
		}
		if isRemoved {
			removed = append(removed, s.Name)
		}
	}
	for _, a := range added {
		updates = append(updates, mantil.ProjectUpdate{
			StaticWebsite: &mantil.StaticWebsiteUpdate{
				Name: a,
			},
			Action: mantil.Add,
		})
	}
	for _, r := range removed {
		updates = append(updates, mantil.ProjectUpdate{
			StaticWebsite: &mantil.StaticWebsiteUpdate{
				Name: r,
			},
			Action: mantil.Remove,
		})
	}
	return updates, nil
}

func (d *DeployCmd) localDirs(path string) ([]string, error) {
	files, err := ioutil.ReadDir(filepath.Join(d.path, path))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	dirs := []string{}
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		dirs = append(dirs, f.Name())
	}
	return dirs, nil
}

func (d *DeployCmd) uploadStaticWebsite(name, bucket string) error {
	log.Info("uploading static website %s to bucket %s", name, bucket)
	basePath := filepath.Join(d.path, StaticSitesDir, name)
	return filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(basePath, path)
		if err != nil {
			return err
		}
		log.Info("uploading file %s...", relPath)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		if err := d.aws.PutObjectToS3Bucket(bucket, relPath, f); err != nil {
			return err
		}
		return nil
	})
}

func (d *DeployCmd) deployRequest() (*mantil.Project, error) {
	type deployReq struct {
		ProjectName string
		Token       string
		Updates     []mantil.ProjectUpdate
	}
	dreq := &deployReq{
		ProjectName: d.project.Name,
		Token:       d.token,
		Updates:     d.updates,
	}
	type deployRsp struct {
		Project *mantil.Project
	}
	dresp := &deployRsp{}
	if err := commands.BackendRequest("deploy", dreq, nil); err != nil {
		return nil, err
	}
	// TODO: temporary fix for api gateway timeout
	type req struct {
		ProjectName string
		Token       string
	}
	r := &req{
		ProjectName: d.project.Name,
		Token:       d.project.Token,
	}
	if err := commands.BackendRequest("data", r, dresp); err != nil {
		return nil, err
	}
	// TODO: temporary fix for obtaining s3 credentials after creating a bucket
	d.refreshCredentials()
	return dresp.Project, nil
}

func (d *DeployCmd) refreshCredentials() error {
	type req struct {
		ProjectName string
		Token       string
	}
	r := &req{
		ProjectName: d.project.Name,
		Token:       d.project.Token,
	}
	creds := &commands.Credentials{}
	if err := commands.BackendRequest("security", r, creds); err != nil {
		return err
	}
	awsClient, err := aws.New(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken)
	if err != nil {
		return err
	}
	d.aws = awsClient
	return nil
}
