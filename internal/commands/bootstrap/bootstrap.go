package bootstrap

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/atoz-technology/mantil-cli/internal/aws"
	"github.com/atoz-technology/mantil-cli/internal/commands"
	"github.com/atoz-technology/mantil.go/pkg/logs"
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

type BootstrapResponse struct {
	APIGatewayURL string
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
	log.Println("Deploying backend infrastructure...")
	req := &BootstrapRequest{
		Destroy: false,
	}
	rsp, err := b.invokeBootstrapLambda(req)
	if err != nil {
		return fmt.Errorf("could not invoke bootstrap function - %v", err)
	}
	config := &commands.BackendConfig{
		APIGatewayURL: rsp.APIGatewayURL,
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("could not save backend config - %v", err)
	}
	return nil
}

func (b *BootstrapCmd) destroy() error {
	req := &BootstrapRequest{
		Destroy: true,
	}
	log.Println("Destroying backend infrastructure...")
	if _, err := b.invokeBootstrapLambda(req); err != nil {
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
	configPath, err := commands.BackendConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(configPath); err != nil {
		return err
	}
	return nil
}

func (b *BootstrapCmd) invokeBootstrapLambda(req *BootstrapRequest) (*BootstrapResponse, error) {
	lambdaARN, err := b.bootstrapLambdaARN()
	if err != nil {
		return nil, err
	}
	l := logs.NewListener()
	wait, err := l.Listen(context.Background(), func(msg string) error {
		fmt.Print(msg)
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer wait()
	clientCtx := map[string]interface{}{
		"custom": map[string]string{
			logs.InboxHeaderKey: l.Subject(),
		},
	}
	rsp := &BootstrapResponse{}
	if err := b.awsClient.InvokeLambdaFunction(lambdaARN, req, rsp, clientCtx); err != nil {
		return nil, fmt.Errorf("could not invoke bootstrap function - %v", err)
	}
	return rsp, nil
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
