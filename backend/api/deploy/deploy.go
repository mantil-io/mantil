package deploy

import (
	"encoding/json"
	"fmt"

	"github.com/mantil-io/mantil/assets"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/backend/log"
	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/terraform"
)

const (
	FunctionsDir = "functions"
	BuildDir     = "build"
)

type Deploy struct {
	aws          *aws.AWS
	projectName  string
	currentState *config.Stage
	desiredState *config.Stage
	tf           *terraform.Terraform
	rc           *config.RuntimeConfig
}

func New(projectName string, desiredState *config.Stage, tf *terraform.Terraform, awsClient *aws.AWS, rc *config.RuntimeConfig) (*Deploy, error) {
	assets.StartServer()
	currentState, err := config.LoadDeploymentState(projectName, desiredState.Name)
	if err != nil {
		currentState = &config.Stage{}
	}
	return &Deploy{
		projectName:  projectName,
		currentState: currentState,
		desiredState: desiredState,
		tf:           tf,
		aws:          awsClient,
		rc:           rc,
	}, nil
}

func (d *Deploy) Deploy() error {
	d.desiredState.AddFunctionDefaults()
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
	return config.SaveDeploymentState(d.projectName, d.desiredState)
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
	bucket, err := config.Bucket(d.aws)
	if err != nil {
		return nil, err
	}
	stage := d.desiredState
	data := terraform.ProjectTemplateData{
		Name:                   d.projectName,
		Bucket:                 bucket,
		BucketPrefix:           config.DeploymentBucketPrefix(d.projectName, stage.Name),
		Functions:              stage.Functions,
		PublicSites:            stage.PublicSites,
		Region:                 d.aws.Region(),
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

func (d *Deploy) terraformDestroy() (*terraform.Terraform, error) {
	bucket, err := config.Bucket(d.aws)
	if err != nil {
		return nil, err
	}
	stage := d.desiredState
	data := terraform.ProjectTemplateData{
		Name:         d.projectName,
		Bucket:       bucket,
		BucketPrefix: config.DeploymentBucketPrefix(d.projectName, stage.Name),
		Region:       d.aws.Region(),
	}
	tf, err := terraform.Project(data)
	if err != nil {
		return nil, err
	}
	return tf, tf.Destroy()
}

func (d *Deploy) updateLambdaFunction(f *config.Function) error {
	log.Info("updating function %s...", f.Name)
	lambdaName := config.ProjectResource(d.projectName, d.desiredState.Name, f.Name)
	var err error
	if f.S3Key != "" {
		var bucket string
		bucket, err = config.Bucket(d.aws)
		if err != nil {
			return err
		}
		err = d.aws.UpdateLambdaFunctionCodeFromS3(lambdaName, bucket, f.S3Key)
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
