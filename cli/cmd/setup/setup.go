package setup

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/mantil-io/mantil.go/pkg/streaming/logs"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/auth"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/build"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/workspace"
)

const (
	lambdaName = "mantil-setup"
)

//go:embed template.yml
var cfTemplate string

type Cmd struct {
	functionsBucket string
	functionsPath   string
	awsClient       *aws.AWS
	accountName     string
	override        bool
}

func New(awsClient *aws.AWS, accountName string, override bool) *Cmd {
	v := build.Version()
	return &Cmd{
		awsClient:       awsClient,
		functionsBucket: v.FunctionsBucket(awsClient.Region()),
		functionsPath:   v.FunctionsPath(),
		accountName:     accountName,
		override:        override,
	}
}

func (c *Cmd) Create() error {
	w, err := workspace.Load()
	if err != nil {
		return err
	}
	if a := w.Account(c.accountName); a != nil {
		msg := fmt.Sprintf("An account named %s already exists in the workspace, please delete it first or use a different name.", c.accountName)
		return log.WithUserMessage(nil, msg)
	}
	ac, err := c.create()
	if err != nil {
		return err
	}
	w.UpsertAccount(ac)
	if err := w.Save(); err != nil {
		return err
	}
	ui.Notice("install successfully finished")
	return nil
}

func (c *Cmd) create() (*workspace.Account, error) {
	if err := c.ensureLambdaExists(); err != nil {
		log.Error(err)
		return nil, err
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("could not create public/private key pair - %v", err)
	}
	ui.Info("Deploying backend infrastructure...")
	log.Printf("invokeLambda functionsBucket: %s, functionsPath: %s, publicKey: %s", c.functionsBucket, c.functionsPath, publicKey)
	rsp, err := c.invokeLambda(&dto.SetupRequest{
		FunctionsBucket: c.functionsBucket,
		FunctionsPath:   c.functionsPath,
		PublicKey:       publicKey,
	})
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	bucketName, err := workspace.Bucket(c.awsClient)
	log.Printf("bucketName: %s", bucketName)
	if err != nil {
		log.Error(err)
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
		log.Error(err)
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
		log.Error(err)
		return err
	}
	return nil
}

func (c *Cmd) isAlreadyRun() (bool, error) {
	log.Printf("lambdaName: %s", lambdaName)
	return c.awsClient.LambdaExists(lambdaName)
}

func (c *Cmd) createLambda() error {
	ui.Info("Creating setup function...")
	td := TemplateData{
		Name:   lambdaName,
		Bucket: c.functionsBucket,
		S3Key:  fmt.Sprintf("%s/setup.zip", c.functionsPath),
		Region: c.awsClient.Region(),
	}
	t, err := renderTemplate(td)
	if err != nil {
		log.Error(err)
		return fmt.Errorf("could not create setup function - %v", err)
	}
	if err := c.awsClient.CreateCloudformationStack(lambdaName, t); err != nil {
		log.Error(err)
		return fmt.Errorf("could not create setup function - %v", err)
	}
	return nil
}

func (c *Cmd) Destroy() error {
	alreadyRun, err := c.isAlreadyRun()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Printf("alreadyRun: %v", alreadyRun)
	if !alreadyRun {
		ui.Errorf("Mantil not found in this account")
		return nil
	}
	if err := c.destroy(); err != nil {
		return err
	}
	if err := workspace.RemoveAccount(c.accountName); err != nil {
		log.Error(err)
		return err
	}
	ui.Notice("infrastructure successfully destroyed")
	return nil
}

func (c *Cmd) destroy() error {
	req := &dto.SetupRequest{
		Destroy: true,
	}
	ui.Info("Destroying backend infrastructure...")
	if _, err := c.invokeLambda(req); err != nil {
		log.Error(err)
		return fmt.Errorf("could not invoke setup function - %v", err)
	}
	ui.Info("Deleting setup function...")
	if err := c.deleteLambda(); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *Cmd) deleteLambda() error {
	if err := c.awsClient.DeleteCloudformationStack(lambdaName); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (c *Cmd) invokeLambda(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	lambdaARN, err := c.lambdaARN()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	l, err := logs.NewNATSListener()
	if err != nil {
		log.Error(err)
		return nil, err
	}
	if err := l.Listen(context.Background(), func(msg string) error {
		ui.Backend(msg)
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
		log.Error(err)
		return nil, fmt.Errorf("could not invoke setup function - %v", err)
	}
	return rsp, nil
}

func (c *Cmd) lambdaARN() (string, error) {
	accountID, err := c.awsClient.AccountID()
	if err != nil {
		log.Error(err)
		return "", err
	}
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		c.awsClient.Region(),
		accountID,
		lambdaName,
	), nil
}

func renderTemplate(data TemplateData) (string, error) {
	tpl := template.Must(template.New("").Parse(cfTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type TemplateData struct {
	Name   string
	Bucket string
	S3Key  string
	Region string
}
