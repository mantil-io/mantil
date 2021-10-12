package setup

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/mantil-io/mantil.go/pkg/streaming/logs"
	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/build"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
)

const (
	resourceNamePrefix = "mantil-setup"
)

//go:embed template.yml
var setupStackTemplate string

type Cmd struct {
	aws             *aws.AWS
	accountName     string
	override        bool // TODO unused
	workspacesStore workspace.Store
	resourceName    string
	resourceTags    map[string]string
}

func New(a *Args) (*Cmd, error) {
	if err := a.validate(); err != nil {
		return nil, log.Wrap(err)
	}
	awsClient, err := a.awsConnect()
	if err != nil {
		return nil, log.WithUserMessage(err, "invalid AWS access credentials")
	}
	wss, err := workspace.NewSingleDeveloperWorkspacesFileStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Cmd{
		aws:             awsClient,
		accountName:     a.AccountName,
		override:        a.Override,
		workspacesStore: wss,
	}, nil
}

func (c *Cmd) Create() error {
	ws, err := c.workspacesStore.LoadOrNew("")
	if err != nil {
		return log.Wrap(err)
	}
	c.resourceName = ws.ResourceName(resourceNamePrefix)
	c.resourceTags = ws.ResourceTags()
	v := build.Version()
	ac, err := ws.NewAccount(c.accountName, c.aws.AccountID(), c.aws.Region(),
		v.FunctionsBucket(c.aws.Region()),
		v.FunctionsPath())
	if err != nil {
		if err == workspace.ErrAccountExists {
			msg := fmt.Sprintf("An account named %s already exists, please delete it first or use a different name.", c.accountName)
			return log.WithUserMessage(nil, msg)
		}
		return log.Wrap(err)
	}

	if err := c.create(ac); err != nil {
		return log.Wrap(err)
	}
	if err := c.workspacesStore.Save(ws); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Cmd) create(ac *workspace.Account) error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	if exists {
		return log.WithUserMessage(nil, "Mantil is already installed in this AWS account")
	}
	ui.Info("==> Installing setup stack...")
	if err := c.createSetupStack(ac.Functions); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Done.\n")
	ui.Info("==> Setting up AWS infrastructure...")
	req := &dto.SetupRequest{
		Bucket:          ac.Bucket,
		FunctionsBucket: ac.Functions.Bucket,
		FunctionsPath:   ac.Functions.Path,
		PublicKey:       ac.Keys.Public,
		ResourceSuffix:  ac.ResourceSuffix(),
		ResourceTags:    c.resourceTags,
	}
	rsp, err := c.invokeLambda(req)
	if err != nil {
		return log.Wrap(err, "failed to invoke setup function")
	}
	ac.Endpoints.Rest = rsp.APIGatewayRestURL
	ac.Endpoints.Ws = rsp.APIGatewayWsURL
	ac.CliRole = rsp.CliRole
	ui.Info("Done.\n")
	return nil
}

func (c *Cmd) backendExists() (bool, error) {
	return c.aws.LambdaExists(c.resourceName)
}

func (c *Cmd) createSetupStack(acf workspace.AccountFunctions) error {
	td := stackTemplateData{
		Name:   c.resourceName,
		Bucket: acf.Bucket,
		S3Key:  fmt.Sprintf("%s/setup.zip", acf.Path),
		Region: c.aws.Region(),
	}
	t, err := renderStackTemplate(td)
	if err != nil {
		return log.Wrap(err, "render template failed")
	}
	if err := c.aws.CloudFormation().CreateStack(c.resourceName, t, c.resourceTags); err != nil {
		return log.Wrap(err, "cloudformation failed")
	}
	// https://github.com/aws-cloudformation/cloudformation-coverage-roadmap/issues/919
	if err := c.aws.TagLogGroup(c.aws.LambdaLogGroup(c.resourceName), c.resourceTags); err != nil {
		return log.Wrap(err, "tagging setup lambda log group failed")
	}
	return nil
}

func (c *Cmd) Destroy() error {
	ws, err := c.workspacesStore.Load("")
	if err != nil {
		return log.Wrap(err)
	}
	c.resourceName = ws.ResourceName(resourceNamePrefix)
	ac := ws.Account(c.accountName)
	if ac == nil {
		return log.WithUserMessage(nil, fmt.Sprintf("Account %s don't exists", c.accountName))
	}

	if err := c.destroy(); err != nil {
		return log.Wrap(err)
	}
	ws.RemoveAccount(ac.Name)
	if err := c.workspacesStore.Save(ws); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Cmd) destroy() error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	if !exists {
		return log.WithUserMessage(nil, "Mantil not found in this AWS account")
	}

	req := &dto.SetupRequest{
		Destroy: true,
	}
	ui.Info("==> Destroying AWS infrastructure...")
	if _, err := c.invokeLambda(req); err != nil {
		return log.Wrap(err, "could not invoke setup function")
	}
	ui.Info("Done.\n")
	ui.Info("==> Removing setup stack...")
	if err := c.aws.CloudFormation().DeleteStack(c.resourceName); err != nil {
		return log.Wrap(err)
	}
	ui.Info("Done.\n")
	return nil
}

func (c *Cmd) invokeLambda(req *dto.SetupRequest) (*dto.SetupResponse, error) {
	log.Printf("invokeLambda %#v", req)
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
	if err := c.aws.InvokeLambdaFunction(c.resourceName, req, rsp, clientCtx); err != nil {
		return nil, log.Wrap(err, "could not invoke setup function")
	}
	return rsp, nil
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
