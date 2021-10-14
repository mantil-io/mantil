package deploy

import (
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

const (
	FunctionsDir     = "functions"
	PublicDir        = "public"
	BuildDir         = "build"
	BinaryName       = "bootstrap"
	DeployHTTPMethod = "deploy"
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
	if err := d.buildAndFindDiffs(); err != nil {
		return log.Wrap(err)
	}
	if !d.HasUpdates() {
		ui.Info("no changes - nothing to deploy")
		return nil
	}
	err := d.callBackend()
	if err != nil {
		return log.Wrap(err)
	}
	if err := workspace.SaveProject(d.ctx.Project, d.ctx.Path); err != nil {
		return log.Wrap(err)
	}

	if d.publicDiff.hasUpdates() {
		if err := d.updatePublicSiteContent(); err != nil {
			return log.Wrap(err)
		}
	}
	ui.Notice("deploy successfully finished")
	return nil
}

func (d *Cmd) HasUpdates() bool {
	return d.functionsDiff.hasUpdates() || d.publicDiff.hasUpdates()
}

func (d *Cmd) infrastructureChanged() bool {
	return d.functionsDiff.infrastructureChanged() || d.publicDiff.infrastructureChanged()
}

func (d *Cmd) buildAndFindDiffs() error {
	fd, err := d.functionUpdates()
	if err != nil {
		return log.Wrap(err)
	}
	d.functionsDiff = fd
	pd, err := d.publicSiteUpdates()
	if err != nil {
		return log.Wrap(err)
	}
	d.publicDiff = pd
	return nil
}

func (d *Cmd) callBackend() error {
	req := &dto.DeployRequest{
		ProjectName:           d.ctx.Project.Name,
		Stage:                 d.ctx.Stage,
		InfrastructureChanged: d.infrastructureChanged(),
		UpdatedFunctions:      d.functionsDiff.updated,
		Account:               d.ctx.Account,
		ResourceSuffix:        d.ctx.Workspace.UID,
		ResourceTags:          d.ctx.ResourceTags(),
	}

	backend, err := d.ctx.Backend()
	if err != nil {
		return log.Wrap(err)
	}
	var rsp dto.DeployResponse
	if err := backend.Call(DeployHTTPMethod, req, &rsp); err != nil {
		return log.Wrap(err)
	}

	if req.InfrastructureChanged {
		d.updateStage(rsp)
	}
	return nil
}

func (d *Cmd) updateStage(rsp dto.DeployResponse) {
	// update stage from response
	s := d.ctx.Stage
	s.SetEndpoints(rsp.Rest, rsp.Ws)
	s.SetPublicBucket(rsp.PublicBucket)
	// TODO: treba li ovo ako sam modificirao stage
	d.ctx.Project.UpsertStage(d.ctx.Stage)

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
