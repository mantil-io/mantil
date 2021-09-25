package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil.go/pkg/streaming/logs"
	"github.com/mantil-io/mantil/internal/auth"
	"github.com/mantil-io/mantil/internal/aws"
	"github.com/mantil-io/mantil/internal/cli/commands"
	"github.com/mantil-io/mantil/internal/cli/log"
)

const (
	setupLambdaName = "mantil-setup"
)

type SetupCmd struct {
	bucket        string
	awsClient     *aws.AWS
	version       string
	functionsPath string
	accountName   string
}

func New(awsClient *aws.AWS, v Version, accountName string) *SetupCmd {
	if accountName == "" {
		accountName = commands.DefaultAccountName
	}
	return &SetupCmd{
		bucket:        v.setupBucket(awsClient.Region()),
		awsClient:     awsClient,
		version:       v.Version,
		functionsPath: v.FunctionsPath,
		accountName:   accountName,
	}
}

type SetupRequest struct {
	Version         string
	FunctionsBucket string
	FunctionsPath   string
	PublicKey       string
	Destroy         bool
}

type SetupResponse struct {
	APIGatewayRestURL string
	APIGatewayWsURL   string
}

func (s *SetupCmd) Create() error {
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
		log.Info("Mantil is already set up on this account, updating credentials and fetching config...")
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		return fmt.Errorf("could not create public/private key pair - %v", err)
	}
	req := &SetupRequest{
		Version:         s.version,
		FunctionsBucket: s.bucket,
		FunctionsPath:   s.functionsPath,
		PublicKey:       publicKey,
	}
	rsp, err := s.invokeSetupLambda(req)
	if err != nil {
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	config, err := commands.LoadWorkspaceConfig()
	if err != nil {
		return fmt.Errorf("could not load workspace config - %v", err)
	}
	config.UpsertAccount(&commands.AccountConfig{
		Name: s.accountName,
		Keys: &commands.AccountKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Endpoints: &commands.AccountEndpoints{
			Rest: rsp.APIGatewayRestURL,
			Ws:   rsp.APIGatewayWsURL,
		},
	})
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
		s.bucket,
		fmt.Sprintf("%s/setup.zip", s.functionsPath),
		[]string{
			fmt.Sprintf("arn:aws:lambda:%s:553035198032:layer:git-lambda2:8", s.awsClient.Region()),
			fmt.Sprintf("arn:aws:lambda:%s:477361877445:layer:terraform-lambda:1", s.awsClient.Region()),
		},
	)
	if err != nil {
		return fmt.Errorf("could not create setup function - %v", err)
	}
	return nil
}

func (s *SetupCmd) Destroy() error {
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
	config, err := commands.LoadWorkspaceConfig()
	if err != nil {
		return err
	}
	config.RemoveAccount(s.accountName)
	if err := config.Save(); err != nil {
		return nil
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
	l, err := logs.NewNATSListener()
	if err != nil {
		return nil, err
	}
	if err := l.Listen(context.Background(), func(msg string) error {
		log.Backend(msg)
		return nil
	}); err != nil {
		return nil, err
	}
	defer l.Wait()
	clientCtx := map[string]interface{}{
		"custom": map[string]string{
			logs.InboxHeaderKey:         l.Subject(),
			logs.StreamingTypeHeaderKey: logs.StreamingTypeNATS,
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
