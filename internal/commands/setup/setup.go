package setup

import (
	"context"
	"fmt"
	"os"

	"github.com/mantil-io/mantil-cli/internal/aws"
	"github.com/mantil-io/mantil-cli/internal/commands"
	"github.com/mantil-io/mantil-cli/internal/log"
	"github.com/mantil-io/mantil.go/pkg/logs"
)

const (
	setupLambdaName = "mantil-setup"
)

type SetupCmd struct {
	awsClient *aws.AWS
}

func New(awsClient *aws.AWS) *SetupCmd {
	return &SetupCmd{
		awsClient: awsClient,
	}
}

type SetupRequest struct {
	Destroy bool
}

type SetupResponse struct {
	APIGatewayURL string
}

func (b *SetupCmd) Setup(destroy bool) error {
	if destroy {
		return b.destroy()
	}
	return b.create()
}

func (b *SetupCmd) create() error {
	setupAlreadyRun, err := b.isSetupAlreadyRun(setupLambdaName)
	if err != nil {
		return err
	}
	if setupAlreadyRun {
		log.Info("setup was already run for this account")
		log.Info("if you wish to recreate the resources first run mantil setup -d to clean up previous setup and then repeat the process")
		return nil
	}

	log.Info("Creating setup function...")
	roleARN, err := b.awsClient.CreateSetupRole(
		setupLambdaName,
		setupLambdaName,
	)
	if err != nil {
		return fmt.Errorf("could not create setup role - %v", err)
	}
	_, err = b.awsClient.CreateLambdaFunction(
		setupLambdaName,
		roleARN,
		"mantil-downloads",
		"functions/setup.zip",
		[]string{
			"arn:aws:lambda:eu-central-1:553035198032:layer:git-lambda2:8",
			"arn:aws:lambda:eu-central-1:477361877445:layer:terraform-lambda:1",
		},
	)
	if err != nil {
		return fmt.Errorf("could not create setup function - %v", err)
	}

	log.Info("Deploying backend infrastructure...")
	req := &SetupRequest{
		Destroy: false,
	}
	rsp, err := b.invokeSetupLambda(req)
	if err != nil {
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	config := &commands.BackendConfig{
		APIGatewayURL: rsp.APIGatewayURL,
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("could not save backend config - %v", err)
	}
	log.Notice("setup successfully finished")
	return nil
}

func (b *SetupCmd) isSetupAlreadyRun(name string) (bool, error) {
	roleExists, err := b.awsClient.RoleExists(name)
	if err != nil {
		return false, err
	}
	lambdaExists, err := b.awsClient.LambdaExists(name)
	if err != nil {
		return false, err
	}
	return roleExists || lambdaExists, nil
}

func (b *SetupCmd) destroy() error {
	setupAlreadyRun, err := b.isSetupAlreadyRun(setupLambdaName)
	if err != nil {
		return err
	}
	if !setupAlreadyRun {
		log.Info("setup function doesn't exist on this account")
		return nil
	}
	req := &SetupRequest{
		Destroy: true,
	}
	log.Info("Destroying backend infrastructure...")
	if _, err := b.invokeSetupLambda(req); err != nil {
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	log.Info("Deleting setup function...")
	if err := b.awsClient.DeleteRole(setupLambdaName); err != nil {
		return err
	}
	if err := b.awsClient.DeletePolicy(setupLambdaName); err != nil {
		return err
	}
	if err := b.awsClient.DeleteLambdaFunction(setupLambdaName); err != nil {
		return err
	}
	configPath, err := commands.BackendConfigPath()
	if err != nil {
		return err
	}
	if err := os.Remove(configPath); err != nil {
		return err
	}
	log.Notice("infrastructure successfully destroyed")
	return nil
}

func (b *SetupCmd) invokeSetupLambda(req *SetupRequest) (*SetupResponse, error) {
	lambdaARN, err := b.setupLambdaARN()
	if err != nil {
		return nil, err
	}
	l := logs.NewListener()
	wait, err := l.Listen(context.Background(), func(msg string) error {
		log.Backend(msg)
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
	rsp := &SetupResponse{}
	if err := b.awsClient.InvokeLambdaFunction(lambdaARN, req, rsp, clientCtx); err != nil {
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	return rsp, nil
}

func (b *SetupCmd) setupLambdaARN() (string, error) {
	accountID, err := b.awsClient.AccountID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		b.awsClient.Region(),
		accountID,
		setupLambdaName,
	), nil
}
