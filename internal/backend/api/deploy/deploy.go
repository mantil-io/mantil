package deploy

import (
	"encoding/json"
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/backend/assets"
	"github.com/mantil-io/mantil/internal/backend/log"
	"github.com/mantil-io/mantil/internal/backend/terraform"
	"github.com/mantil-io/mantil/internal/mantil"
	"github.com/mantil-io/mantil/internal/util"
)

const (
	FunctionsDir = "functions"
	BuildDir     = "build"
)

type Deploy struct {
	aws     *aws.AWS
	project *mantil.Project
	stage   *mantil.Stage
	tf      *terraform.Terraform
}

func New(project *mantil.Project, stage *mantil.Stage, tf *terraform.Terraform) (*Deploy, error) {
	awsClient, err := aws.New()
	if err != nil {
		return nil, err
	}
	assets.StartServer()
	return &Deploy{
		aws:     awsClient,
		project: project,
		stage:   stage,
		tf:      tf,
	}, nil
}

func (d *Deploy) Deploy() error {
	d.stage.AddFunctionDefaults()
	infrastructureChanged, err := d.processUpdates()
	if err != nil {
		return err
	}
	// apply terraform if there are changes in infrastructure
	if infrastructureChanged {
		log.Info("applying changes to infrastructure...")
		if err := d.applyInfrastructure(); err != nil {
			return err
		}
	}
	return mantil.SaveProjectS3(d.project)
}

func (d *Deploy) processUpdates() (bool, error) {
	oldStage := d.project.Stage(d.stage.Name)
	if oldStage == nil {
		d.project.Stages = append(d.project.Stages, d.stage)
		return true, nil
	}
	d.project.UpsertStage(d.stage)
	if funcsAddedOrRemoved(oldStage, d.stage) {
		return true, nil
	}
	if sitesAddedOrRemoved(oldStage, d.stage) {
		return true, nil
	}
	if err := d.updateFunctions(oldStage, d.stage); err != nil {
		return false, err
	}
	return false, nil
}

func funcsAddedOrRemoved(oldStage, newStage *mantil.Stage) bool {
	var oldFuncs, newFuncs []string
	for _, f := range oldStage.Functions {
		oldFuncs = append(oldFuncs, f.Name)
	}
	for _, f := range newStage.Functions {
		newFuncs = append(newFuncs, f.Name)
	}
	return addedOrRemoved(oldFuncs, newFuncs)
}

func sitesAddedOrRemoved(oldStage, newStage *mantil.Stage) bool {
	var oldSites, newSites []string
	for _, s := range oldStage.PublicSites {
		oldSites = append(oldSites, s.Name)
	}
	for _, s := range newStage.PublicSites {
		newSites = append(newSites, s.Name)
	}
	return addedOrRemoved(oldSites, newSites)
}

func addedOrRemoved(old, new []string) bool {
	if removed := util.DiffArrays(old, new); len(removed) > 0 {
		return true
	}
	if added := util.DiffArrays(new, old); len(added) > 0 {
		return true
	}
	return false
}

func (d *Deploy) updateFunctions(oldStage, newStage *mantil.Stage) error {
	for _, f := range newStage.Functions {
		for _, of := range oldStage.Functions {
			if f.Name != of.Name {
				continue
			}
			if f.Hash != of.Hash {
				if err := d.updateLambdaFunction(f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *Deploy) applyInfrastructure() error {
	tf := d.tf
	if err := tf.ApplyForProject(d.project, d.stage.Name, d.aws, false); err != nil {
		return fmt.Errorf("could not apply terraform for project %s - %v", d.project.Name, err)
	}
	url, err := tf.Output("url", true)
	if err != nil {
		return fmt.Errorf("could not read terraform output variable for api url - %v", err)
	}
	wsUrl, err := tf.Output("ws_url", true)
	if err != nil {
		return fmt.Errorf("could not read terraform output variable for api ws url - %v", err)
	}
	sites, err := tf.Output("static_websites", false)
	if err != nil {
		return fmt.Errorf("coult not read terraform output variable for static websites - %v", err)
	}
	if err := d.updateWebsitesConfig(sites); err != nil {
		return err
	}
	d.stage.Endpoints = &mantil.StageEndpoints{
		Rest: url,
		Ws:   wsUrl,
	}
	return nil
}

func (d *Deploy) updateLambdaFunction(f *mantil.Function) error {
	log.Info("updating function %s...", f.Name)
	lambdaName := mantil.ProjectResource(d.project.Name, d.stage.Name, f.Name)
	var err error
	if f.S3Key != "" {
		err = d.aws.UpdateLambdaFunctionCodeFromS3(lambdaName, d.project.Bucket, f.S3Key)
	} else {
		err = fmt.Errorf("could not update lambda function %s due to missing key", lambdaName)
	}
	if err != nil {
		return err
	}
	log.Debug("waiting for function's update status to be successful...")
	return d.aws.WaitLambdaFunctionUpdated(lambdaName)
}

func (d *Deploy) updateWebsitesConfig(tfOutput string) error {
	type sitesOutput struct {
		Name   string `json:"name"`
		Bucket string `json:"bucket"`
	}
	os := &[]sitesOutput{}
	if err := json.Unmarshal([]byte(tfOutput), os); err != nil {
		return err
	}
	for _, o := range *os {
		for _, s := range d.stage.PublicSites {
			if o.Name == s.Name {
				s.Bucket = o.Bucket
			}
		}
	}
	return nil
}
