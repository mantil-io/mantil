package deploy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/api/log"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

type Deploy struct {
	req           *dto.DeployRequest
	stage         *workspace.Stage
	awsClient     *aws.AWS
	publicBuckets map[string]string
}

func New() *Deploy {
	return &Deploy{}
}

func (d *Deploy) Invoke(ctx context.Context, req *dto.DeployRequest) (*dto.DeployResponse, error) {
	if err := d.init(req); err != nil {
		return nil, err
	}
	// TODO zasto je ovo na ovoj strani
	d.stage.AddFunctionDefaults()
	if err := d.deploy(); err != nil {
		return nil, err
	}
	if err := workspace.SaveStageStage(d.req.Account.Bucket, d.req.ProjectName, d.stage); err != nil {
		return nil, err
	}
	return &dto.DeployResponse{
		Rest:          d.stage.Endpoints.Rest,
		Ws:            d.stage.Endpoints.Ws,
		PublicBuckets: d.publicBuckets,
	}, nil
}

func (d *Deploy) init(req *dto.DeployRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	d.req = req
	d.stage = d.req.Stage
	d.awsClient = awsClient
	return nil
}

func (d *Deploy) deploy() error {
	if d.req.InfrastructureChanged {
		log.Info("applying changes to infrastructure...")
		return d.applyInfrastructure()
	}
	return d.updateFunctions()
}

func (d *Deploy) applyInfrastructure() error {
	tf, err := d.terraformCreate()
	if err != nil {
		return fmt.Errorf("could not apply terraform for project %s - %v", d.req.ProjectName, err)
	}
	// TODO terrafrom prikuplja outpute u Outputs, nema potrebe pokretiati ga ponovo za svaki
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
	d.stage.Endpoints = &workspace.StageEndpoints{
		Rest: url,
		Ws:   wsUrl,
	}
	return nil
}

func (d *Deploy) terraformCreate() (*terraform.Terraform, error) {
	stage := d.stage
	data := terraform.ProjectTemplateData{
		Name:                   d.req.ProjectName,
		Bucket:                 d.req.Account.Bucket,
		BucketPrefix:           workspace.StageBucketPrefix(d.req.ProjectName, stage.Name),
		Functions:              stage.Functions,
		Public:                 stage.Public,
		Region:                 d.awsClient.Region(),
		Stage:                  stage.Name,
		RuntimeFunctionsBucket: d.req.Account.Functions.Bucket,
		RuntimeFunctionsPath:   d.req.Account.Functions.Path,
		GlobalEnv:              workspace.StageEnv(d.req.ProjectName, stage.Name),
		ResourceTags:           d.req.ResourceTags,
	}
	tf, err := terraform.Project(data)
	if err != nil {
		return nil, err
	}
	return tf, tf.Create()
}

func (d *Deploy) updateFunctions() error {
	for _, fn := range d.req.UpdatedFunctions {
		for _, f := range d.stage.Functions {
			if fn == f.Name {
				if err := d.updateLambdaFunction(f); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (d *Deploy) updateLambdaFunction(f *workspace.Function) error {
	log.Info("updating function %s...", f.Name)
	lambdaName := workspace.ProjectResource(d.req.ProjectName, d.stage.Name, f.Name)
	var err error
	if f.S3Key != "" {
		err = d.awsClient.UpdateLambdaFunctionCodeFromS3(lambdaName, d.req.Account.Bucket, f.S3Key)
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
	d.publicBuckets = make(map[string]string)
	for _, o := range *os {
		for _, s := range d.stage.Public {
			if o.Name == s.Name {
				s.Bucket = o.Bucket
			}
			d.publicBuckets[o.Name] = o.Bucket
		}
	}
	return nil
}
