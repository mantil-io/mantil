package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil.go/pkg/streaming/logs"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/workspace"
)

const (
	lambdaName = "mantil-setup"
)

type Cmd struct {
	functionsBucket string
	functionsPath   string
	awsClient       *aws.AWS
	accountName     string
	override        bool
}

func New(awsClient *aws.AWS, v *VersionInfo, accountName string, override bool) *Cmd {
	return &Cmd{
		functionsBucket: v.functionsBucket(awsClient.Region()),
		awsClient:       awsClient,
		functionsPath:   v.functionsPath(),
		accountName:     accountName,
		override:        override,
	}
}

func (c *Cmd) Create() error {
	ac, err := c.create()
	if err != nil {
		return err
	}
	if err = workspace.UpsertAccount(ac); err != nil {
		return err
	}
	log.Notice("install successfully finished")
	return nil
}

func (c *Cmd) create() (*workspace.Account, error) {
	if err := c.ensureLambdaExists(); err != nil {
		log.Printf("[ERROR] %w", err)
		return nil, err
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		log.Printf("[ERROR] %w", err)
		return nil, fmt.Errorf("could not create public/private key pair - %v", err)
	}
	log.Info("Deploying backend infrastructure...")
	log.Printf("invokeLambda functionsBucket: %s, functionsPath: %s, publicKey: %s", c.functionsBucket, c.functionsPath, publicKey)
	rsp, err := c.invokeLambda(&dto.SetupRequest{
		FunctionsBucket: c.functionsBucket,
		FunctionsPath:   c.functionsPath,
		PublicKey:       publicKey,
	})
	if err != nil {
		log.Printf("[ERROR] %w", err)
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	bucketName, err := workspace.Bucket(c.awsClient)
	log.Printf("bucketName: %s", bucketName)
	if err != nil {
		log.Printf("[ERROR] %w", err)
		return nil, err
	}
	return &workspace.Account{
		Name:   c.accountName,
		Bucket: bucketName,
		Keys: &workspace.AccountKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Endpoints: &workspace.AccountEndpoints{
			Rest: rsp.APIGatewayRestURL,
			Ws:   rsp.APIGatewayWsURL,
		},
	}, nil
}

func (c *Cmd) ensureLambdaExists() error {
	alreadyRun, err := c.isAlreadyRun()
	if err != nil {
		log.Printf("[ERROR]: %w", err)
		return err
	}
	log.Printf("alreadyRun: %v override: %v", alreadyRun, c.override)
	if alreadyRun {
		if c.override {
			return nil
		}
		err := fmt.Errorf("Mantil is already installed use override flag if you want to change acccess tokens")
		log.Printf("Mantil already installed and override is not set returning: %s", err)
		return err
	}
	if err := c.createLambda(); err != nil {
		log.Printf("[ERROR] %w", err)
		return err
	}
	return nil
}

func (c *Cmd) isAlreadyRun() (bool, error) {
	log.Printf("lambdaName: %s", lambdaName)
	return c.awsClient.LambdaExists(lambdaName)
}

func (c *Cmd) createLambda() error {
	log.Info("Creating setup function...")
	if err := c.awsClient.CreateCloudformationStack(lambdaName, CloudformationTemplate); err != nil {
		log.Printf("[ERROR] %w", err)
		return fmt.Errorf("could not create setup function - %v", err)
	}
	return nil
}

func (c *Cmd) Destroy() error {
	alreadyRun, err := c.isAlreadyRun()
	if err != nil {
		log.Printf("[ERROR] %w", err)
		return err
	}
	log.Printf("alreadyRun: %v", alreadyRun)
	if !alreadyRun {
		log.Errorf("Mantil not found in this account")
		return nil
	}
	if err := c.destroy(); err != nil {
		return err
	}
	if err := workspace.RemoveAccount(c.accountName); err != nil {
		log.Printf("[ERROR] %w", err)
		return err
	}
	log.Notice("infrastructure successfully destroyed")
	return nil
}

func (c *Cmd) destroy() error {
	req := &dto.SetupRequest{
		Destroy: true,
	}
	log.Info("Destroying backend infrastructure...")
	if _, err := c.invokeLambda(req); err != nil {
		log.Printf("[ERROR] %w", err)
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	log.Info("Deleting setup function...")
	if err := c.deleteLambda(); err != nil {
		log.Printf("[ERROR] %w", err)
		return err
	}
	return nil
}

func (c *Cmd) deleteLambda() error {
	return c.awsClient.DeleteCloudformationStack(lambdaName)

}

func (c *Cmd) invokeLambda(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	lambdaARN, err := c.lambdaARN()
	if err != nil {
		log.Printf("[ERROR] %w", err)
		return nil, err
	}
	l, err := logs.NewNATSListener()
	if err != nil {
		log.Printf("[ERROR] %w", err)
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
	rsp := &dto.SetupResponse{}
	if err := c.awsClient.InvokeLambdaFunction(lambdaARN, req, rsp, clientCtx); err != nil {
		log.Printf("[ERROR] %w", err)
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	return rsp, nil
}

func (c *Cmd) lambdaARN() (string, error) {
	accountID, err := c.awsClient.AccountID()
	if err != nil {
		log.Printf("[ERROR] %w", err)
		return "", err
	}
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		c.awsClient.Region(),
		accountID,
		lambdaName,
	), nil
}
