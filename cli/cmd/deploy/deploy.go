package deploy

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

const (
	FunctionsDir = "functions"
	PublicDir    = "public"
	BuildDir     = "build"
	BinaryName   = "bootstrap"
)

type Args struct {
	Stage string
}

type Cmd struct {
	ctx           *project.Context
	awsClient     *aws.AWS
	functionsDiff resourceDiff
	publicDiff    resourceDiff
}

func New(a Args) (*Cmd, error) {
	ctx, err := project.NewContext()
	if err != nil {
		return nil, log.Wrap(err)
	}
	stage := ctx.ResolveStage(a.Stage)
	if stage == nil {
		return nil, log.WithUserMessage(nil, "The specified stage doesn't exist, create it with `mantil stage new`.")
	}
	ctx.SetStage(stage)
	awsClient, err := ctx.AWSClient()
	if err != nil {
		return nil, log.Wrap(err)
	}
	d := &Cmd{
		ctx:       ctx,
		awsClient: awsClient,
	}
	return d, nil
}

func NewFromContext(ctx *project.Context) (*Cmd, error) {
	awsClient, err := ctx.AWSClient()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Cmd{
		ctx:       ctx,
		awsClient: awsClient,
	}, nil
}

func (d *Cmd) Deploy() error {
	ui.Info("deploying stage %s to account %s", d.ctx.Stage.Name, d.ctx.Account.Name)
	if err := d.deploySync(); err != nil {
		return err
	}
	if !d.HasUpdates() {
		ui.Info("no changes - nothing to deploy")
		return nil
	}
	p, err := d.deployRequest()
	if err != nil {
		return err
	}
	if err := workspace.SaveProject(p, d.ctx.Path); err != nil {
		return err
	}
	ui.Notice("deploy successfully finished")
	if err := d.updatePublicSiteContent(); err != nil {
		return err
	}
	return nil
}

func (d *Cmd) HasUpdates() bool {
	return d.functionsDiff.hasUpdates() || d.publicDiff.hasUpdates()
}

func (d *Cmd) InfrastructureChanged() bool {
	return d.functionsDiff.infrastructureChanged() || d.publicDiff.infrastructureChanged()
}

func (d *Cmd) deploySync() error {
	fd, err := d.functionUpdates()
	if err != nil {
		return err
	}
	d.functionsDiff = fd
	pd, err := d.publicSiteUpdates()
	if err != nil {
		return err
	}
	d.publicDiff = pd
	return nil
}

func (d *Cmd) localDirs(path string) ([]string, error) {
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

const DeployHTTPMethod = "deploy"

func (d *Cmd) deployRequest() (*workspace.Project, error) {
	req := &dto.DeployRequest{
		ProjectName:           d.ctx.Project.Name,
		Stage:                 d.ctx.Stage,
		InfrastructureChanged: d.InfrastructureChanged(),
		UpdatedFunctions:      d.functionsDiff.updated,
		Account:               d.ctx.Account,
		ResourceTags:          d.ctx.ResourceTags(),
	}

	b, err := d.ctx.Backend()
	if err != nil {
		return nil, err
	}
	var rsp dto.DeployResponse
	if err := b.Call(DeployHTTPMethod, req, &rsp); err != nil {
		return nil, err
	}

	// TODO: temporary fix for api gateway timeout
	dreq := &dto.DataRequest{
		Bucket:      d.ctx.Account.Bucket,
		ProjectName: d.ctx.Project.Name,
		StageName:   d.ctx.Stage.Name,
	}
	dresp := &dto.DataResponse{}
	if err := d.ctx.RuntimeRequest("data", dreq, dresp, false); err != nil {
		return nil, err
	}
	d.ctx.Stage = dresp.Stage
	d.ctx.Project.UpsertStage(d.ctx.Stage)
	// TODO: temporary fix for obtaining s3 credentials after creating a bucket
	d.refreshCredentials()
	return d.ctx.Project, nil
}

func (d *Cmd) refreshCredentials() error {
	awsClient, err := d.ctx.AWSClient()
	if err != nil {
		return err
	}
	d.awsClient = awsClient
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

type resourceDiff struct {
	added   []string
	removed []string
	updated []string
}

func (d *resourceDiff) infrastructureChanged() bool {
	return len(d.added) > 0 || len(d.removed) > 0
}

func (d *resourceDiff) hasUpdates() bool {
	return d.infrastructureChanged() || len(d.updated) > 0
}
