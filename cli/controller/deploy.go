package controller

import (
	"fmt"
	"time"

	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

const (
	FunctionsDir     = "functions"
	PublicDir        = "public"
	BuildDir         = "build"
	BinaryName       = "bootstrap"
	DeployHTTPMethod = "deploy"
	HashCharacters   = 8
)

type DeployArgs struct {
	Stage string
}

type Deploy struct {
	repoPut func(bucket, key string, content []byte) error
	diff    *domain.StageDiff

	store *domain.FileStore
	stage *domain.Stage
	title string

	buildDuration     time.Duration
	lastBuildDuration time.Duration
	uploadDuration    time.Duration
	uploadBytes       int64
	updateDuration    time.Duration
}

func NewDeploy(a DeployArgs) (*Deploy, error) {
	fs, stage, err := newStoreWithStage(a.Stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return NewDeployWithStage(fs, stage)
}

func NewDeployWithStage(fs *domain.FileStore, stage *domain.Stage) (*Deploy, error) {
	d := &Deploy{
		store: fs,
		stage: stage,
	}
	if err := d.setAWSclient(); err != nil {
		return nil, log.Wrap(err)
	}
	return d, nil
}

func (d *Deploy) setAWSclient() error {
	stage := d.stage
	awsClient, err := awsClient(stage.Node(), stage.Project(), stage)
	if err != nil {
		return log.Wrap(err)
	}
	d.repoPut = awsClient.S3().Put
	return nil
}

func (d *Deploy) DeployWithTitle(title string) error {
	d.title = title
	return d.Deploy()
}

func (d *Deploy) Deploy() error {
	if d.title == "" {
		d.title = fmt.Sprintf("Building and deploying %s to stage %s", d.stage.Project().Name, d.stage.Name)
	}
	ui.Title("\n%s\n", d.title)
	if err := d.deploy(); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Build time: %v, upload: %v (%s), update: %v",
		d.buildDuration.Round(time.Millisecond),
		d.uploadDuration.Round(time.Millisecond),
		formatFileSizeUnits(d.uploadBytes),
		d.updateDuration.Round(time.Millisecond))

	// create deploy event
	de := domain.Deploy{
		InfrastructureChanged: d.diff.InfrastructureChanged(),
		BuildDuration:         toMS(d.buildDuration),
		UploadDuration:        toMS(d.uploadDuration),
		UploadBytes:           int(d.uploadBytes),
		UpdateDuration:        toMS(d.updateDuration),
	}
	a, u, r := d.diff.FunctionsAddedUpdatedRemoved()
	de.Functions.Added = a
	de.Functions.Updated = u
	de.Functions.Removed = r
	log.Event(domain.Event{Deploy: &de})
	return nil
}

func (d *Deploy) deploy() error {
	ui.Info("Building...")
	if err := d.buildAndFindDiffs(); err != nil {
		return log.Wrap(err)
	}
	if !d.HasUpdates() {
		ui.Info("No changes - nothing to deploy")
		return nil
	}
	if len(d.diff.UpdatedFunctions()) > 0 {
		ui.Info("Uploading changes...")
		if err := d.uploadTimer(func() error { return d.uploadFunctions() }); err != nil {
			return log.Wrap(err)
		}
	}

	if d.diff.InfrastructureChanged() {
		ui.Title("Setting up AWS infrastructure...\n")
	} else {
		ui.Info("Updating infrastructure...")
	}
	err := d.updateTimer(func() error { return d.callBackend() })
	if err != nil {
		return log.Wrap(err)
	}

	if d.diff.HasPublicUpdates() {
		if err := d.setAWSclient(); err != nil {
			return log.Wrap(err)
		}
		ui.Info("Updating public content...")
		if err := d.uploadTimer(func() error { return d.updatePublicContent() }); err != nil {
			return log.Wrap(err)
		}
	}

	if err := d.store.Store(); err != nil {
		return log.Wrap(err)
	}
	ui.Info("")
	ui.Title("Deploy successful!\n")
	return nil
}

func (d *Deploy) HasUpdates() bool {
	return d.diff.HasUpdates()
}

func (d *Deploy) buildAndFindDiffs() error {
	lf, err := d.localFunctions()
	if err != nil {
		return log.Wrap(err)
	}
	var ph string
	if d.hasPublic() {
		ph, err = d.publicHash()
		if err != nil {
			return log.Wrap(err)
		}
	}
	diff, err := d.stage.ApplyChanges(lf, ph)
	if err != nil {
		return log.Wrap(err)
	}
	d.diff = diff
	return nil
}

func (d *Deploy) callBackend() error {
	ni, err := nodeInvoker(d.stage.Node())
	if err != nil {
		return log.Wrap(err)
	}
	var rsp dto.DeployResponse
	if err := ni.Do(DeployHTTPMethod, d.backendRequest(), &rsp); err != nil {
		return log.Wrap(err)
	}
	if d.diff.InfrastructureChanged() {
		d.updateStage(rsp)
	}
	return nil
}

func (d *Deploy) backendRequest() dto.DeployRequest {
	req := dto.DeployRequest{
		NodeBucket:         d.stage.Node().Bucket,
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
			Project:             d.stage.Project().Name,
			Bucket:              d.stage.Node().Bucket,
			BucketPrefix:        d.stage.StateBucketPrefix(),
			Functions:           fns,
			Region:              d.stage.Node().Region,
			Stage:               d.stage.Name,
			NodeFunctionsBucket: d.stage.Node().Functions.Bucket,
			NodeFunctionsPath:   d.stage.Node().Functions.Path,
			ResourceSuffix:      d.stage.Node().ResourceSuffix(),
			ResourceTags:        d.stage.ResourceTags(),
			WsEnv:               d.stage.WsEnv(),
			HasPublic:           d.stage.HasPublic(),
		}
	}
	return req
}

func (d *Deploy) workspaceFunction2dto(w domain.Function) dto.Function {
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

func (d *Deploy) updateStage(rsp dto.DeployResponse) {
	d.stage.SetEndpoints(rsp.Rest, rsp.Ws)
	if rsp.PublicBucket != "" {
		d.stage.SetPublicBucket(rsp.PublicBucket)
	}
}

func (d *Deploy) buildTimer(cb func() error) error {
	before := d.buildDuration
	err := timer(&d.buildDuration, cb)
	d.lastBuildDuration = d.buildDuration - before
	return err
}

func (d *Deploy) uploadTimer(cb func() error) error {
	return timer(&d.uploadDuration, cb)
}

func (d *Deploy) updateTimer(cb func() error) error {
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

func toMS(d time.Duration) int {
	return int(d / time.Millisecond)
}
