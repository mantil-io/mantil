package deploy

import (
	"fmt"
	"time"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/controller"
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
	awsClient *aws.AWS
	diff      *workspace.StageDiff

	store *workspace.FileStore
	stage *workspace.Stage
	path  string

	buildDuration  time.Duration
	uploadDuration time.Duration
	uploadBytes    int64
	updateDuration time.Duration
}

func New(a Args) (*Cmd, error) {
	fs, err := controller.NewStoreWithStage(a.Stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return NewWithStage(fs, fs.Stage(a.Stage))
}

func NewWithStage(fs *workspace.FileStore, stage *workspace.Stage) (*Cmd, error) {
	d := &Cmd{
		store: fs,
		stage: stage,
		path:  fs.ProjectRoot(),
	}
	if err := d.setAWSclient(); err != nil {
		return nil, err
	}
	return d, nil
}

func (d *Cmd) setAWSclient() error {
	stage := d.stage
	awsClient, err := controller.AWSClient(stage.Account(), stage.Project(), stage)
	if err != nil {
		return log.Wrap(err)
	}
	d.awsClient = awsClient
	return nil
}

func (d *Cmd) Deploy() error {
	if err := d.deploy(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (d *Cmd) deploy() error {
	ui.Info("==> Building...")
	if err := d.buildAndFindDiffs(); err != nil {
		return log.Wrap(err)
	}
	ui.Info("")
	if !d.HasUpdates() {
		ui.Info("No changes - nothing to deploy")
		return nil
	}
	if len(d.diff.UpdatedFunctions()) > 0 {
		ui.Info("==> Uploading...")
		if err := d.uploadTimer(func() error { return d.uploadFunctions() }); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
	}

	if d.diff.HasFunctionUpdates() {
		if d.diff.InfrastructureChanged() {
			ui.Info("==> Setting up AWS infrastructure...")
		} else {
			ui.Info("==> Updating...")
		}
		err := d.updateTimer(func() error { return d.callBackend() })
		if err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
	}

	if d.diff.HasPublicUpdates() {
		if err := d.setAWSclient(); err != nil {
			return log.Wrap(err)
		}
		ui.Info("==> Updating public content...")
		if err := d.uploadTimer(func() error { return d.updatePublicSiteContent() }); err != nil {
			return log.Wrap(err)
		}
		ui.Info("")
	}

	if err := d.store.Store(); err != nil {
		return log.Wrap(err)
	}

	ui.Info("Build time: %v, upload: %v (%s), update: %v",
		d.buildDuration.Round(time.Millisecond),
		d.uploadDuration.Round(time.Millisecond),
		formatFileSizeUnits(d.uploadBytes),
		d.updateDuration.Round(time.Millisecond))
	return nil
}

func (d *Cmd) HasUpdates() bool {
	return d.diff.HasUpdates()
}

func (d *Cmd) buildAndFindDiffs() error {
	lf, err := d.localFunctions()
	if err != nil {
		return log.Wrap(err)
	}
	lp, err := d.localPublicSites()
	if err != nil {
		return log.Wrap(err)
	}
	diff, err := d.stage.ApplyChanges(lf, lp)
	if err != nil {
		return log.Wrap(err)
	}
	d.diff = diff
	return nil
}

func (d *Cmd) callBackend() error {
	backend, err := controller.Backend(d.stage.Account())
	if err != nil {
		return log.Wrap(err)
	}
	var rsp dto.DeployResponse
	if err := backend.Call(DeployHTTPMethod, d.backendRequest(), &rsp); err != nil {
		return log.Wrap(err)
	}
	if d.diff.InfrastructureChanged() {
		d.updateStage(rsp)
	}
	return nil
}

func (d *Cmd) backendRequest() dto.DeployRequest {
	req := dto.DeployRequest{
		AccountBucket:      d.stage.Account().Bucket,
		FunctionsForUpdate: nil,
		StageTemplate:      nil,
	}
	var fns []dto.Function
	var fnsu []dto.Function
	for _, f := range d.stage.Functions {
		df := d.workspaceFunction2dto(*f)
		fns = append(fns, df)
		for _, fn := range d.diff.UpdatedFunctions() {
			if fn == f.Name {
				fnsu = append(fnsu, df)
			}
		}
	}
	req.FunctionsForUpdate = fnsu
	if d.diff.InfrastructureChanged() {
		req.StageTemplate = &dto.StageTemplate{
			Project:                d.stage.Project().Name,
			Bucket:                 d.stage.Account().Bucket,
			BucketPrefix:           d.stage.BucketPrefix(),
			Functions:              fns,
			Region:                 d.stage.Account().Region,
			Stage:                  d.stage.Name,
			AccountFunctionsBucket: d.stage.Account().Functions.Bucket,
			AccountFunctionsPath:   d.stage.Account().Functions.Path,
			ResourceSuffix:         d.stage.Account().ResourceSuffix(),
			ResourceTags:           d.stage.ResourceTags(),
		}
	}
	return req
}

func (d *Cmd) workspaceFunction2dto(w workspace.Function) dto.Function {
	return dto.Function{
		Name:       w.Name,
		LambdaName: w.LambdaName(),
		S3Key:      w.S3Key,
		Runtime:    "provided.al2",
		Handler:    "bootstrap",
		MemorySize: w.MemorySize,
		Timeout:    w.Timeout,
		Env:        w.Env,
	}
}

func (d *Cmd) updateStage(rsp dto.DeployResponse) {
	d.stage.SetEndpoints(rsp.Rest, rsp.Ws)
	d.stage.SetPublicBucket(rsp.PublicBucket)
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
func formatFileSizeUnits(b int64) string {
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
