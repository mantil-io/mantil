package bootstrap

import (
	"fmt"
	"log"

	"github.com/atoz-technology/mantil-cli/internal/aws"
)

const (
	bootstrapLambdaName = "mantil-backend-lambda-bootstrap"
)

type BootstrapCmd struct {
	awsClient *aws.AWS
}

func New(awsClient *aws.AWS) *BootstrapCmd {
	return &BootstrapCmd{
		awsClient: awsClient,
	}
}

type BootstrapRequest struct {
	Destroy bool
}

func (b *BootstrapCmd) Bootstrap(destroy bool) error {
	if destroy {
		return b.destroy()
	}
	return b.create()
}

func (b *BootstrapCmd) create() error {
	log.Println("Creating bootstrap function...")
	roleARN, err := b.awsClient.CreateBootstrapRole(
		bootstrapLambdaName,
		bootstrapLambdaName,
	)
	if err != nil {
		return fmt.Errorf("could not create bootstrap role - %v", err)
	}
	_, err = b.awsClient.CreateLambdaFunction(
		bootstrapLambdaName,
		roleARN,
		"mantil-downloads",
		"functions/bootstrap.zip",
		[]string{
			"arn:aws:lambda:eu-central-1:553035198032:layer:git-lambda2:8",
			"arn:aws:lambda:eu-central-1:477361877445:layer:terraform-lambda:1",
		},
	)
	if err != nil {
		return fmt.Errorf("could not create bootstrap function - %v", err)
	}
	lambdaARN, err := b.bootstrapLambdaARN()
	if err != nil {
		return err
	}
	log.Println("Deploying backend infrastructure...")
	req := &BootstrapRequest{
		Destroy: false,
	}
	if err := b.awsClient.InvokeLambdaFunction(lambdaARN, req, nil); err != nil {
		return fmt.Errorf("could not invoke bootstrap function - %v", err)
	}
	return nil
}

func (b *BootstrapCmd) destroy() error {
	req := &BootstrapRequest{
		Destroy: true,
	}
	lambdaARN, err := b.bootstrapLambdaARN()
	if err != nil {
		return err
	}
	log.Println("Destroying backend infrastructure...")
	if err := b.awsClient.InvokeLambdaFunction(lambdaARN, req, nil); err != nil {
		return fmt.Errorf("could not invoke bootstrap function - %v", err)
	}
	log.Println("Deleting bootstrap function...")
	if err := b.awsClient.DeleteRole(bootstrapLambdaName); err != nil {
		return err
	}
	if err := b.awsClient.DeletePolicy(bootstrapLambdaName); err != nil {
		return err
	}
	if err := b.awsClient.DeleteLambdaFunction(bootstrapLambdaName); err != nil {
		return err
	}
	return nil
}

func (b *BootstrapCmd) bootstrapLambdaARN() (string, error) {
	accountID, err := b.awsClient.AccountID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		b.awsClient.Region(),
		accountID,
		bootstrapLambdaName,
	), nil
}
