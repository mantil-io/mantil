package setup

import (
	"context"
	"fmt"

	"github.com/mantil-io/mantil.go/pkg/streaming/logs"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/mantil/commands"
	"github.com/mantil-io/mantil/cli/mantil/log"
)

const (
	resourceName = "mantil-setup"
)

type Cmd struct {
	bucket        string
	awsClient     *aws.AWS
	version       string
	functionsPath string
	accountName   string
	cleanupSteps  []cleanupStep
}

func New(awsClient *aws.AWS, v Version, accountName string) *Cmd {
	if accountName == "" {
		accountName = commands.DefaultAccountName
	}
	return &Cmd{
		bucket:        v.setupBucket(awsClient.Region()),
		awsClient:     awsClient,
		version:       v.String(), // TODO: sto ce mi ovaj version kada se nigdje ne koristi
		functionsPath: v.functionsPath(),
		accountName:   accountName,
	}
}

func (c *Cmd) Create() error {
	ac, err := c.createWithCleanup()
	if err != nil {
		return err
	}
	if err = commands.WorkspaceUpsertAccount(ac); err != nil {
		return err
	}
	log.Notice("setup successfully finished")
	return nil
}

func (c *Cmd) createWithCleanup() (*commands.AccountConfig, error) {
	ac, err := c.create()
	if err != nil {
		c.cleanup()
		return nil, err
	}
	return ac, nil
}

func (c *Cmd) create() (*commands.AccountConfig, error) {
	if err := c.ensureLambdaExists(); err != nil {
		return nil, err
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("could not create public/private key pair - %v", err)
	}
	log.Info("Deploying backend infrastructure...")
	rsp, err := c.invokeLambda(&dto.SetupRequest{
		// TODO: sto ce mi ovaj version kada se nigdje ne koristi
		Version:         c.version,
		FunctionsBucket: c.bucket,
		FunctionsPath:   c.functionsPath,
		PublicKey:       publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	return &commands.AccountConfig{
		Name: c.accountName,
		Keys: &commands.AccountKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Endpoints: &commands.AccountEndpoints{
			Rest: rsp.APIGatewayRestURL,
			Ws:   rsp.APIGatewayWsURL,
		},
	}, nil
}

func (c *Cmd) ensureLambdaExists() error {
	alreadyRun, err := c.isAlreadyRun()
	if err != nil {
		return err
	}
	if alreadyRun {
		// BUG: ovo u nastavku ne valja promjenit ce credentials pa padaju svi postojeci
		log.Info("Mantil is already set up on this account, updating credentials and fetching config...")
		return nil
	}
	if err := c.createLambda(); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) isAlreadyRun() (bool, error) {
	return c.awsClient.LambdaExists(resourceName)
}

func (c *Cmd) createLambda() error {
	log.Info("Creating setup function...")
	roleARN, err := c.createSetupRole()
	if err != nil {
		return err
	}
	_, err = c.awsClient.CreateLambdaFunction(
		resourceName,
		roleARN,
		c.bucket,
		fmt.Sprintf("%s/setup.zip", c.functionsPath),
		[]string{
			fmt.Sprintf("arn:aws:lambda:%s:553035198032:layer:git-lambda2:8", c.awsClient.Region()),
			fmt.Sprintf("arn:aws:lambda:%s:477361877445:layer:terraform-lambda:1", c.awsClient.Region()),
		},
	)
	if err != nil {
		return fmt.Errorf("could not create setup function - %v", err)
	}
	c.newCleanupStep("Setup lambda function", func() error {
		return c.awsClient.DeleteLambdaFunction(resourceName)
	})
	return nil
}

func (c *Cmd) createSetupRole() (string, error) {
	r, err := c.awsClient.CreateRole(resourceName, setupAssumeRolePolicy())
	if err != nil {
		return "", err
	}
	c.newCleanupStep("Setup IAM Role", func() error {
		return c.awsClient.DeleteRole(resourceName)
	})
	p, err := c.awsClient.CreatePolicy(resourceName, setupLambdaPolicy(*r.RoleId, resourceName))
	if err != nil {
		return "", err
	}
	c.newCleanupStep("Setup IAM Policy", func() error {
		return c.awsClient.DeletePolicy(resourceName)
	})
	if err := c.awsClient.AttachRolePolicy(*p.Arn, *r.RoleName); err != nil {
		return "", err
	}
	return *r.Arn, nil
}

func setupAssumeRolePolicy() string {
	return `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Action": "sts:AssumeRole",
				"Principal": {
					"Service": "lambda.amazonaws.com"
				},
				"Effect": "Allow"
			}
		]
	}`
}

func setupLambdaPolicy(roleID, lambdaName string) string {
	return `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Resource": "*",
				"Action": "*"
			},
			{
				"Effect": "Deny",
				"Resource": "*",
				"Action": "*",
				"Condition": {
					"StringNotLike": {
						"aws:userid": "` + roleID + `:` + lambdaName + `"
					}
				}
			}
		]
	}`
}

func (c *Cmd) deleteSetupRole() error {
	if err := c.awsClient.DeleteRole(resourceName); err != nil {
		return err
	}
	if err := c.awsClient.DeletePolicy(resourceName); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) Destroy() error {
	alreadyRun, err := c.isAlreadyRun()
	if err != nil {
		return err
	}
	if !alreadyRun {
		log.Errorf("setup function doesn't exist on this account")
		return nil
	}
	if err := c.destroy(); err != nil {
		return err
	}
	if err := commands.WorkspaceRemoveAccount(c.accountName); err != nil {
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
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	log.Info("Deleting setup function...")
	if err := c.deleteLambda(); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) deleteLambda() error {
	if err := c.deleteSetupRole(); err != nil {
		return err
	}
	if err := c.awsClient.DeleteLambdaFunction(resourceName); err != nil {
		return err
	}
	return nil
}

func (c *Cmd) invokeLambda(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	lambdaARN, err := c.lambdaARN()
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
	rsp := &dto.SetupResponse{}
	if err := c.awsClient.InvokeLambdaFunction(lambdaARN, req, rsp, clientCtx); err != nil {
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	return rsp, nil
}

func (c *Cmd) lambdaARN() (string, error) {
	accountID, err := c.awsClient.AccountID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		c.awsClient.Region(),
		accountID,
		resourceName,
	), nil
}

type cleanupStep struct {
	name    string
	cleanup func() error
}

func (c *Cmd) newCleanupStep(name string, cleanup func() error) {
	c.cleanupSteps = append(c.cleanupSteps, cleanupStep{name, cleanup})
}

func (c *Cmd) cleanup() {
	log.Errorf("encountered error, cleaning up resources...")
	for i, step := range c.cleanupSteps {
		if err := step.cleanup(); err != nil {
			c.logLeftoverCleanupSteps(c.cleanupSteps[i:])
			break
		}
	}
}

func (c *Cmd) logLeftoverCleanupSteps(cs []cleanupStep) {
	log.Errorf("error recovery failed - some of the resources could not be cleaned up:")
	for _, s := range cs {
		log.Errorf(s.name)
	}
}
