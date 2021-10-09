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
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

const (
	lambdaName = "mantil-setup"
)

//go:embed template.yml
var setupStackTemplate string

type Cmd struct {
	functionsBucket string
	functionsPath   string
	awsClient       *aws.AWS
	accountName     string
	override        bool
}

func New(a *Args) (*Cmd, error) {
	if err := a.validate(); err != nil {
		return nil, log.Wrap(err)
	}
	awsClient, err := a.awsConnect()
	if err != nil {
		return nil, log.WithUserMessage(err, "invalid AWS access credentials")
	}
	v := build.Version()
	return &Cmd{
		awsClient:       awsClient,
		functionsBucket: v.FunctionsBucket(awsClient.Region()),
		functionsPath:   v.FunctionsPath(),
		accountName:     a.AccountName,
		override:        a.Override,
	}, nil
}

func (c *Cmd) Create() error {
	w, err := workspace.Load()
	if err != nil {
		return log.Wrap(err)
	}
	if a := w.Account(c.accountName); a != nil {
		msg := fmt.Sprintf("An account named %s already exists, please delete it first or use a different name.", c.accountName)
		return log.WithUserMessage(nil, msg)
	}
	ac, err := c.create()
	if err != nil {
		return log.Wrap(err)
	}
	w.UpsertAccount(ac)
	if err := w.Save(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Cmd) create() (*workspace.Account, error) {
	if err := c.ensureLambdaExists(); err != nil {
		return nil, log.Wrap(err)
	}
	publicKey, privateKey, err := auth.CreateKeyPair()
	if err != nil {
		return nil, log.Wrap(err, "could not create public/private key pair")
	}
	ui.Info("==> Setting up AWS infrastructure...")
	log.Printf("invokeLambda functionsBucket: %s, functionsPath: %s, publicKey: %s", c.functionsBucket, c.functionsPath, publicKey)
	rsp, err := c.invokeLambda(&dto.SetupRequest{
		FunctionsBucket: c.functionsBucket,
		FunctionsPath:   c.functionsPath,
		PublicKey:       publicKey,
	})
	if err != nil {
		return nil, log.Wrap(err, "failed to invoke setup function")
	}
	ui.Info("Done.\n")
	return &workspace.Account{
		Name:   c.accountName,
		ID:     c.awsClient.AccountID(),
		Region: c.awsClient.Region(),
		Bucket: workspace.Bucket(c.awsClient),
		Keys: workspace.AccountKeys{
			Public:  publicKey,
			Private: privateKey,
		},
		Endpoints: workspace.AccountEndpoints{
			Rest: rsp.APIGatewayRestURL,
			Ws:   rsp.APIGatewayWsURL,
		},
		Functions: workspace.AccountFunctions{
			Bucket: c.functionsBucket,
			Path:   c.functionsPath,
		},
	}, nil
}

func (c *Cmd) ensureLambdaExists() error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	log.Printf("exists: %v override: %v", exists, c.override)
	if exists {
		if c.override {
			return nil
		}
		return log.WithUserMessage(nil, "Mantil is already installed in this AWS account.\nUse override flag if you want to change acccess tokens.")
	}
	if err := c.createSetupStack(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Cmd) backendExists() (bool, error) {
	return c.awsClient.LambdaExists(lambdaName)
}

func (c *Cmd) createSetupStack() error {
	ui.Info("==> Installing setup stack...")
	td := stackTemplateData{
		Name:   lambdaName,
		Bucket: c.functionsBucket,
		S3Key:  fmt.Sprintf("%s/setup.zip", c.functionsPath),
		Region: c.awsClient.Region(),
	}
	t, err := renderStackTemplate(td)
	if err != nil {
		return log.Wrap(err, "render template failed")
	}
	if err := c.awsClient.CreateCloudformationStack(lambdaName, t); err != nil {
		return log.Wrap(err, "cloudformation failed")
	}
	ui.Info("Done.\n")
	return nil
}

func (c *Cmd) Destroy() error {
	alreadyRun, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	log.Printf("alreadyRun: %v", alreadyRun)
	if !alreadyRun {
		return log.WithUserMessage(nil, "Mantil not found in this account")
	}
	if err := c.destroy(); err != nil {
		return log.Wrap(err)
	}
	if err := workspace.RemoveAccount(c.accountName); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Cmd) destroy() error {
	req := &dto.SetupRequest{
		Destroy: true,
	}
	ui.Info("==> Destroying AWS infrastructure...")
	if _, err := c.invokeLambda(req); err != nil {
		return log.Wrap(err, "could not invoke setup function")
	}
	ui.Info("Done.\n")
	ui.Info("==> Removing setup stack...")
	if err := c.deleteLambda(); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Done.\n")
	return nil
}

func (c *Cmd) deleteLambda() error {
	if err := c.awsClient.DeleteCloudformationStack(lambdaName); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Cmd) invokeLambda(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	lambdaARN, err := c.lambdaARN()
	if err != nil {
		return nil, log.Wrap(err)
	}
	l, err := logs.NewNATSListener()
	if err != nil {
		return nil, log.Wrap(err)
	}
	tp := terraform.NewLogParser()
	if err := l.Listen(context.Background(), func(line string) error {
		log.Printf(line)
		// TODO if !ok it is not terraform line
		if l, ok := tp.Parse(line); ok && l != "" {
			ui.Info(l)
		}
		return nil
	}); err != nil {
		return nil, log.Wrap(err)
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
		return nil, log.Wrap(err, "could not invoke setup function")
	}
	return rsp, nil
}

func (c *Cmd) lambdaARN() (string, error) {
	return fmt.Sprintf(
		"arn:aws:lambda:%s:%s:function:%s",
		c.awsClient.Region(),
		c.awsClient.AccountID(),
		lambdaName,
	), nil
}

func renderStackTemplate(data stackTemplateData) (string, error) {
	tpl := template.Must(template.New("").Parse(setupStackTemplate))
	buf := bytes.NewBuffer(nil)
	if err := tpl.Execute(buf, data); err != nil {
		return "", log.Wrap(err)
	}
	return buf.String(), nil
}

type stackTemplateData struct {
	Name   string
	Bucket string
	S3Key  string
	Region string
}
