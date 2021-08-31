package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil.go/pkg/logs"
	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/log"
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

func (s *SetupCmd) Setup(destroy bool) error {
	if destroy {
		return s.destroy()
	}
	return s.create()
}

func (s *SetupCmd) create() error {
	setupAlreadyRun, err := s.isSetupAlreadyRun()
	if err != nil {
		return err
	}
	if !setupAlreadyRun {
		if err := s.firstTimeSetup(); err != nil {
			return err
		}
		log.Info("Deploying backend infrastructure...")
	} else {
		log.Info("Mantil is already set up on this account, fetching config...")
	}
	req := &SetupRequest{
		Destroy: false,
	}
	rsp, err := s.invokeSetupLambda(req)
	if err != nil {
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	config := &commands.BackendConfig{
		APIGatewayURL: rsp.APIGatewayURL,
	}
	if err := commands.CreateConfigDir(); err != nil {
		return fmt.Errorf("could not create config directory - %v", err)
	}
	if err := config.Save(); err != nil {
		return fmt.Errorf("could not save backend config - %v", err)
	}
	log.Notice("setup successfully finished")
	return nil
}

func (s *SetupCmd) isSetupAlreadyRun() (bool, error) {
	setupLambdaExists, err := s.awsClient.LambdaExists(setupLambdaName)
	if err != nil {
		return false, err
	}
	return setupLambdaExists, nil
}

func (s *SetupCmd) firstTimeSetup() error {
	log.Info("Creating setup function...")
	roleARN, err := s.awsClient.CreateSetupRole(
		setupLambdaName,
		setupLambdaName,
	)
	if err != nil {
		return fmt.Errorf("could not create setup role - %v", err)
	}
	_, err = s.awsClient.CreateLambdaFunction(
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
	return nil
}

func (s *SetupCmd) destroy() error {
	setupAlreadyRun, err := s.isSetupAlreadyRun()
	if err != nil {
		return err
	}
	if !setupAlreadyRun {
		log.Errorf("setup function doesn't exist on this account")
		return nil
	}
	req := &SetupRequest{
		Destroy: true,
	}
	log.Info("Destroying backend infrastructure...")
	if _, err := s.invokeSetupLambda(req); err != nil {
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	log.Info("Deleting setup function...")
	if err := s.awsClient.DeleteRole(setupLambdaName); err != nil {
		return err
	}
	if err := s.awsClient.DeletePolicy(setupLambdaName); err != nil {
		return err
	}
	if err := s.awsClient.DeleteLambdaFunction(setupLambdaName); err != nil {
		return err
	}
	if err := commands.RemoveConfigDir(); err != nil {
		return fmt.Errorf("could not remove config directory - %v", err)
	}
	log.Notice("infrastructure successfully destroyed")
	return nil
}

func (s *SetupCmd) setupLambdaExists() (bool, error) {
	exists, err := s.awsClient.LambdaExists(setupLambdaName)
	return exists, err
}

func (s *SetupCmd) invokeSetupLambda(req *SetupRequest) (*SetupResponse, error) {
	lambdaARN, err := s.setupLambdaARN()
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
	if err := s.awsClient.InvokeLambdaFunction(lambdaARN, req, rsp, clientCtx); err != nil {
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	return rsp, nil
}

func (s *SetupCmd) setupLambdaARN() (string, error) {
	accountID, err := s.awsClient.AccountID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		s.awsClient.Region(),
		accountID,
		setupLambdaName,
	), nil
}
