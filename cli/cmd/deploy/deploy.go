package deploy

import (
	"github.com/mantil-io/mantil/cli/cmd/project"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

const (
	FunctionsDir   = "functions"
	PublicSitesDir = "public"
	BuildDir       = "build"
	BinaryName     = "bootstrap"
)

type DeployCmd struct {
	aws                *aws.AWS
	ctx                *project.Context
	updatedPublicSites []string
}

func New(ctx *project.Context, awsClient *aws.AWS) (*DeployCmd, error) {
	d := &DeployCmd{
		aws: awsClient,
		ctx: ctx,
	}
	return d, nil
}

func (d *DeployCmd) Deploy() (bool, error) {
	log.UI.Info("deploying stage %s to account %s", d.ctx.Stage.Name, d.ctx.Account.Name)
	updated, err := d.deploySync()
	if err != nil {
		return false, err
	}
	if !updated {
		log.UI.Info("no changes - nothing to deploy")
		return false, nil
	}
	p, err := d.deployRequest()
	if err != nil {
		return false, err
	}
	if err := workspace.SaveProject(p, d.ctx.Path); err != nil {
		return false, err
	}
	log.UI.Notice("deploy successfully finished")
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
	files, err := ioutil.ReadDir(filepath.Join(d.ctx.Path, path))
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

func (d *DeployCmd) deployRequest() (*workspace.Project, error) {
	type deployReq struct {
		ProjectName string
		Stage       *workspace.Stage
	}
	dreq := &deployReq{
		ProjectName: d.ctx.Project.Name,
		Stage:       d.ctx.Stage,
	}
	if err := d.ctx.RuntimeRequest("deploy", dreq, nil, true); err != nil {
		return nil, err
	}
	// TODO: temporary fix for api gateway timeout
	type req struct {
		ProjectName string
		StageName   string
	}
	r := &req{
		ProjectName: d.ctx.Project.Name,
		StageName:   d.ctx.Stage.Name,
	}
	type dataResp struct {
		Stage *workspace.Stage
	}
	dresp := &dataResp{}
	if err := d.ctx.RuntimeRequest("data", r, dresp, false); err != nil {
		return nil, err
	}
	d.ctx.Stage = dresp.Stage
	d.ctx.Project.UpsertStage(d.ctx.Stage)
	// TODO: temporary fix for obtaining s3 credentials after creating a bucket
	d.refreshCredentials()
	return d.ctx.Project, nil
}

func (d *DeployCmd) refreshCredentials() error {
	awsClient, err := d.ctx.InitialiseAWSSDK()
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
