package deploy

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/log"
	"github.com/mantil-io/mantil/internal/mantil"
)

const (
	FunctionsDir   = "functions"
	StaticSitesDir = "public"
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
	d.project = p
	log.Notice("deploy successfully finished")
	if p.ApiURL != d.config.ApiURL {
		d.config.ApiURL = p.ApiURL
		if err = d.config.Save(d.path); err != nil {
			return err
		}
	}
	return d.updateStaticWebsiteContents()
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
	if err := d.validateUpdates(); err != nil {
		return fmt.Errorf("deployment failed - %v", err)
	}
	return nil
}

func (d *DeployCmd) HasUpdates() bool {
	return len(d.updates) > 0
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

func (d *DeployCmd) validateUpdates() error {
	for _, u := range d.updates {
		if u.Function != nil && u.Action == mantil.Add && !mantil.FunctionNameAvailable(u.Function.Name) {
			return fmt.Errorf("api name \"%s\" is reserved", u.Function.Name)
		}
	}
	return nil
}
