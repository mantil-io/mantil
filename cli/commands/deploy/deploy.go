package deploy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/commands"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/config"
)

const (
	FunctionsDir   = "functions"
	PublicSitesDir = "public"
	BuildDir       = "build"
	BinaryName     = "bootstrap"
)

type DeployCmd struct {
	aws                *aws.AWS
	account            *commands.AccountConfig
	project            *config.Project
	stage              *config.Stage
	path               string
	updatedPublicSites []string
}

func New(account *commands.AccountConfig, project *config.Project, stage *config.Stage, awsClient *aws.AWS, path string) (*DeployCmd, error) {
	d := &DeployCmd{
		aws:     awsClient,
		account: account,
		project: project,
		stage:   stage,
		path:    path,
	}
	return d, nil
}

func (d *DeployCmd) Deploy() (bool, error) {
	updated, err := d.deploySync()
	if err != nil {
		return false, err
	}
	if !updated {
		log.Info("no changes - nothing to deploy")
		return false, nil
	}
	p, err := d.deployRequest()
	if err != nil {
		return false, err
	}
	if err := config.SaveProject(p, d.path); err != nil {
		return false, err
	}
	log.Notice("deploy successfully finished")
	if err := d.updatePublicSiteContent(); err != nil {
		return false, err
	}
	return true, nil
}

func (d *DeployCmd) deploySync() (updated bool, err error) {
	functionsUpdated, err := d.functionUpdates()
	if err != nil {
		return false, err
	}
	updatedSites, err := d.publicSiteUpdates()
	if err != nil {
		return false, err
	}
	d.updatedPublicSites = updatedSites
	return functionsUpdated || len(updatedSites) > 0, nil
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

func (d *DeployCmd) deployRequest() (*config.Project, error) {
	type deployReq struct {
		ProjectName string
		Stage       *config.Stage
	}
	dreq := &deployReq{
		ProjectName: d.project.Name,
		Stage:       d.stage,
	}
	if err := commands.BackendRequest("deploy", dreq, nil, true); err != nil {
		return nil, err
	}
	// TODO: temporary fix for api gateway timeout
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: d.project.Name,
		StageName:   d.stage.Name,
	}
	type dataResp struct {
		Stage *config.Stage
	}
	dresp := &dataResp{}
	if err := commands.BackendRequest("data", r, dresp, false); err != nil {
		return nil, err
	}
	d.stage = dresp.Stage
	d.project.UpsertStage(d.stage)
	// TODO: temporary fix for obtaining s3 credentials after creating a bucket
	d.refreshCredentials()
	return d.project, nil
}

func (d *DeployCmd) refreshCredentials() error {
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: d.project.Name,
		StageName:   d.stage.Name,
	}
	creds := &commands.Credentials{}
	if err := commands.BackendRequest("security", r, creds, false); err != nil {
		return err
	}
	awsClient, err := aws.NewWithCredentials(creds.AccessKeyID, creds.SecretAccessKey, creds.SessionToken, creds.Region)
	if err != nil {
		return err
	}
	d.aws = awsClient
	return nil
}

// returns a1 - a2
func diffArrays(a1 []string, a2 []string) []string {
	m := make(map[string]bool)
	for _, e := range a2 {
		m[e] = true
	}
	var diff []string
	for _, e := range a1 {
		if m[e] {
			continue
		}
		diff = append(diff, e)
	}
	return diff
}

// returns a1 n a2
func intersectArrays(a1 []string, a2 []string) []string {
	m := make(map[string]bool)
	for _, e := range a1 {
		m[e] = true
	}
	var intersection []string
	for _, e := range a2 {
		if m[e] {
			intersection = append(intersection, e)
		}
	}
	return intersection
}