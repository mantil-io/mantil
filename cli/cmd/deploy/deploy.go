package deploy

import (
	"fmt"
	"time"

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
	configChanged bool

	functionsForUpload []uploadData
	buildDuration      time.Duration
	uploadDuration     time.Duration
	uploadBytes        int64
	updateDuration     time.Duration
}

type uploadData struct {
	name       string
	binaryPath string
	s3Key      string
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
	if err := d.deploy(); err != nil {
		return log.WithUserMessage(err, "Failed")
	}
	return nil
}

func (d *Cmd) deploy() error {
	ui.Info("==> Building...")
	if err := d.buildAndFindDiffs(); err != nil {
		return log.Wrap(err)
	}
	ui.Info("")
	if err := d.applyConfiguration(); err != nil {
		return log.Wrap(err)
	}
	if !d.HasUpdates() {
		ui.Info("No changes - nothing to deploy")
		return nil
	}
	if len(d.functionsForUpload) > 0 {
		ui.Info("==> Uploading...")
		if err := d.uploadTimer(func() error { return d.upload() }); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
	}

	if d.hasFunctionUpdates() {
		if d.infrastructureChanged() {
			ui.Info("==> Setting up AWS infrastructure...")
		} else {
			ui.Info("==> Updating...")
		}
		err := d.updateTimer(func() error { return d.callBackend() })
		if err != nil {
			return log.Wrap(err)
		}
		if err := workspace.SaveProject(d.ctx.Project, d.ctx.Path); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
	}

	if d.publicDiff.hasUpdates() {
		ui.Info("==> Updating public content...")
		if err := d.uploadTimer(func() error { return d.updatePublicSiteContent() }); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
	}

	ui.Info("Build time: %v, upload: %v (%s), update: %v",
		d.buildDuration.Round(time.Millisecond),
		d.uploadDuration.Round(time.Millisecond),
		byteCountIEC(d.uploadBytes),
		d.updateDuration.Round(time.Millisecond))
	return nil
}

func (d *Cmd) applyConfiguration() error {
	d.ctx.Stage.AddFunctionDefaults()
	envChanged, err := d.ctx.Stage.ApplyEnv(
		d.ctx.Path,
		d.ctx.Project.Name,
		d.ctx.Workspace.UID,
	)
	if err != nil {
		return log.Wrap(err)
	}
	d.configChanged = envChanged
	return nil
}

func (d *Cmd) HasUpdates() bool {
	return d.functionsDiff.hasUpdates() ||
		d.publicDiff.hasUpdates() ||
		d.configChanged
}

func (d *Cmd) hasFunctionUpdates() bool {
	return d.functionsDiff.hasUpdates() ||
		d.configChanged
}

func (d *Cmd) infrastructureChanged() bool {
	return d.functionsDiff.infrastructureChanged() ||
		d.publicDiff.infrastructureChanged() ||
		d.configChanged
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

func (d *Cmd) buildTimer(cb func() error) error {
	return timer(&d.buildDuration, cb)
}

func (d *Cmd) uploadTimer(cb func() error) error {
	return timer(&d.uploadDuration, cb)
}

func (d *Cmd) updateTimer(cb func() error) error {
	return timer(&d.updateDuration, cb)
}

func timer(dur *time.Duration, cb func() error) error {
	start := time.Now()
	defer func() {
		*dur += time.Now().Sub(start)
	}()
	if err := cb(); err != nil {
		return err
	}
	return nil
}

// stolen from: https://yourbasic.org/golang/formatting-byte-size-to-human-readable-format/
// ubuntu units policy: https://wiki.ubuntu.com/UnitsPolicy
func byteCountIEC(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB",
		float64(b)/float64(div), "KMGTPE"[exp])
}
