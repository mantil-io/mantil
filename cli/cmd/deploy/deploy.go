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
	backend, err := d.ctx.Backend()
	if err != nil {
		return log.Wrap(err)
	}
	var rsp dto.DeployResponse
	if err := backend.Call(DeployHTTPMethod, d.backendRequest(), &rsp); err != nil {
		return log.Wrap(err)
	}
	if d.infrastructureChanged() {
		d.updateStage(rsp)
	}
	return nil
}

func (d *Cmd) backendRequest() dto.DeployRequest {
	// TODO remove for new projects, done in AddFunction
	d.ctx.Stage.AddFunctionDefaults()

	req := dto.DeployRequest{
		AccountBucket:      d.ctx.Account.Bucket,
		FunctionsForUpdate: nil,
		StageTemplate:      nil,
	}
	var fns []dto.Function
	var fnsu []dto.Function
	for _, f := range d.ctx.Stage.Functions {
		df := d.workspaceFunction2dto(*f)
		fns = append(fns, df)
		for _, fn := range d.functionsDiff.updated {
			if fn == f.Name {
				fnsu = append(fnsu, df)
			}
		}
	}
	req.FunctionsForUpdate = fnsu
	if d.infrastructureChanged() {
		req.StageTemplate = &dto.StageTemplate{
			Project:                d.ctx.Project.Name,
			Bucket:                 d.ctx.Account.Bucket,
			BucketPrefix:           workspace.StageBucketPrefix(d.ctx.Project.Name, d.ctx.Stage.Name),
			Functions:              fns,
			Region:                 d.ctx.Account.Region,
			Stage:                  d.ctx.Stage.Name,
			AccountFunctionsBucket: d.ctx.Account.Functions.Bucket,
			AccountFunctionsPath:   d.ctx.Account.Functions.Path,
			ResourceSuffix:         d.ctx.Workspace.UID,
			GlobalEnv:              workspace.StageEnv(d.ctx.Project.Name, d.ctx.Stage.Name, d.ctx.Workspace.UID),
			ResourceTags:           d.ctx.ResourceTags(),
		}
	}
	return req
}

func (d *Cmd) workspaceFunction2dto(w workspace.Function) dto.Function {
	lambdaName := workspace.ProjectResource(d.ctx.Project.Name, d.ctx.Stage.Name, w.Name, d.ctx.Workspace.UID)
	return dto.Function{
		Name:       w.Name,
		LambdaName: lambdaName,
		S3Key:      w.S3Key,
		Runtime:    w.Runtime,
		Handler:    w.Handler,
		MemorySize: w.MemorySize,
		Timeout:    w.Timeout,
		Env:        w.Env,
	}
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
