package deploy

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/kit/aws"
	"github.com/mantil-io/mantil/node/dto"
	"github.com/mantil-io/mantil/node/terraform"
)

type Deploy struct {
	req       dto.DeployRequest
	rsp       dto.DeployResponse
	awsClient *aws.AWS
}

func New() *Deploy {
	return &Deploy{}
}

func (d *Deploy) Invoke(ctx context.Context, req dto.DeployRequest) (*dto.DeployResponse, error) {
	ok, _ := domain.IsAuthorizedForProject(ctx, req.StageTemplate.Project)
	if !ok {
		return nil, domain.ErrNotAuthorized
	}
	if err := d.init(req); err != nil {
		return nil, err
	}
	if err := d.deploy(); err != nil {
		return nil, err
	}
	return &d.rsp, nil
}

func (d *Deploy) init(req dto.DeployRequest) error {
	awsClient, err := aws.New()
	if err != nil {
		return fmt.Errorf("error initializing aws client - %w", err)
	}
	d.req = req
	d.awsClient = awsClient
	return nil
}

func (d *Deploy) deploy() error {
	if d.req.StageTemplate != nil {
		return d.applyInfrastructure()
	}
	return d.updateFunctions()
}

func (d *Deploy) applyInfrastructure() error {
	if d.req.StageTemplate == nil {
		return nil
	}
	// call terraform
	tf, err := d.terraformCreate()
	if err != nil {
		return err
	}
	// collect terraform output
	d.rsp.Rest, err = tf.Output("url")
	if err != nil {
		return err
	}
	d.rsp.Ws, err = tf.Output("ws_url")
	if err != nil {
		return err
	}
	if d.req.StageTemplate.HasPublic {
		d.rsp.PublicBucket, err = tf.Output("public_site_bucket")
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *Deploy) terraformCreate() (*terraform.Terraform, error) {
	tf, err := terraform.Project(*d.req.StageTemplate)
	if err != nil {
		return nil, fmt.Errorf("terrafrom.Project failed %w,", err)
	}
	return tf, tf.Create()
}

func (d *Deploy) updateFunctions() error {
	for _, fn := range d.req.FunctionsForUpdate {
		if err := d.updateLambdaFunction(fn); err != nil {
			return err
		}
	}
	return nil
}

func (d *Deploy) updateLambdaFunction(f dto.Function) error {
	err := d.awsClient.UpdateLambdaFunctionCodeFromS3(f.LambdaName, d.req.NodeBucket, f.S3Key)
	if err != nil {
		return err
	}
	return d.awsClient.WaitLambdaFunctionUpdated(f.LambdaName)
}
