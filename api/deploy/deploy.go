package deploy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mantil-io/mantil/api/log"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

type Deploy struct {
	projectName  string
	currentState *config.Stage
	desiredState *config.Stage
	rc           *config.RuntimeConfig
	bucketName   string
	awsClient    *aws.AWS
}

type DeployRequest struct {
	ProjectName string
	Stage       *config.Stage
}

type DeployResponse struct{}

func New() *Deploy {
	return &Deploy{}
}

func (d *Deploy) Invoke(ctx context.Context, req *DeployRequest) (*DeployResponse, error) {
	if err := d.init(req); err != nil {
		return nil, err
	}
	return d.deploy()
}

func (d *Deploy) init(req *DeployRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	currentState, err := config.LoadDeploymentState(req.ProjectName, req.Stage.Name)
	if err != nil {
		return fmt.Errorf("error fetching current deployment state - %w", err)
	}
	rc, err := config.LoadRuntimeConfig(awsClient)
	if err != nil {
		return fmt.Errorf("error fetching runtime config - %w", err)
	}
	bucketName, err := config.Bucket(awsClient)
	if err != nil {
		return fmt.Errorf("error fetching bucket name - %w", err)
	}
	d.projectName = req.ProjectName
	d.desiredState = req.Stage
	d.currentState = currentState
	d.rc = rc
	d.bucketName = bucketName
	d.awsClient = awsClient
	return nil
}

func (d *Deploy) deploy() (*DeployResponse, error) {
	d.desiredState.AddFunctionDefaults()
	infrastructureChanged, err := d.processUpdates()
	if err != nil {
		return nil, err
	}
	if infrastructureChanged {
		log.Info("applying changes to infrastructure...")
		if err := d.applyInfrastructure(); err != nil {
			return nil, err
		}
	}
	return nil, config.SaveDeploymentState(d.projectName, d.desiredState)
}

func (d *Deploy) processUpdates() (bool, error) {
	if funcsAddedOrRemoved(d.currentState, d.desiredState) {
		return true, nil
	}
	if sitesAddedOrRemoved(d.currentState, d.desiredState) {
		return true, nil
	}
	if err := d.updateFunctions(d.currentState, d.desiredState); err != nil {
		return false, err
	}
	return false, nil
}

func funcsAddedOrRemoved(current, new *config.Stage) bool {
	var oldFuncs, newFuncs []string
	for _, f := range current.Functions {
		oldFuncs = append(oldFuncs, f.Name)
	}
	for _, f := range new.Functions {
		newFuncs = append(newFuncs, f.Name)
	}
	return addedOrRemoved(oldFuncs, newFuncs)
}

func sitesAddedOrRemoved(current, new *config.Stage) bool {
	var oldSites, newSites []string
	for _, s := range current.PublicSites {
		oldSites = append(oldSites, s.Name)
	}
	for _, s := range new.PublicSites {
		newSites = append(newSites, s.Name)
	}
	return addedOrRemoved(oldSites, newSites)
}

func addedOrRemoved(current, new []string) bool {
	if removed := diffArrays(current, new); len(removed) > 0 {
		return true
	}
	if added := diffArrays(new, current); len(added) > 0 {
		return true
	}
	return false
}

func (d *Deploy) updateFunctions(oldStage, newStage *config.Stage) error {
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
	tf, err := d.terraformCreate()
	if err != nil {
		return fmt.Errorf("could not apply terraform for project %s - %v", d.projectName, err)
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
	d.desiredState.Endpoints = &config.StageEndpoints{
		Rest: url,
		Ws:   wsUrl,
	}
	return nil
}

func (d *Deploy) terraformCreate() (*terraform.Terraform, error) {
	stage := d.desiredState
	data := terraform.ProjectTemplateData{
		Name:                   d.projectName,
		Bucket:                 d.bucketName,
		BucketPrefix:           config.DeploymentBucketPrefix(d.projectName, stage.Name),
		Functions:              stage.Functions,
		PublicSites:            stage.PublicSites,
		Region:                 d.awsClient.Region(),
		Stage:                  stage.Name,
		RuntimeFunctionsBucket: d.rc.FunctionsBucket,
		RuntimeFunctionsPath:   d.rc.FunctionsPath,
	}
	tf, err := terraform.Project(data)
	if err != nil {
		return nil, err
	}
	return tf, tf.Create()
}

func (d *Deploy) updateLambdaFunction(f *config.Function) error {
	log.Info("updating function %s...", f.Name)
	lambdaName := config.ProjectResource(d.projectName, d.desiredState.Name, f.Name)
	var err error
	if f.S3Key != "" {
		err = d.awsClient.UpdateLambdaFunctionCodeFromS3(lambdaName, d.bucketName, f.S3Key)
	} else {
		err = fmt.Errorf("could not update lambda function %s due to missing key", lambdaName)
	}
	if err != nil {
		return err
	}
	log.Debug("waiting for function's update status to be successful...")
	return d.awsClient.WaitLambdaFunctionUpdated(lambdaName)
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
		for _, s := range d.desiredState.PublicSites {
			if o.Name == s.Name {
				s.Bucket = o.Bucket
			}
		}
	}
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
