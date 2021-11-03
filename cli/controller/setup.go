package controller

import (
	_ "embed"
	"fmt"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/backend"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

//go:embed setup_stack_template.yml
var setupStackTemplate string

const (
	stackResourceCount = 4
)

type Setup struct {
	aws          *aws.AWS
	accountName  string
	override     bool // TODO unused
	store        *domain.FileStore
	resourceTags map[string]string
	stackName    string
	lambdaName   string
}

type stackTemplateData struct {
	Name   string
	Bucket string
	S3Key  string
	Region string
}

func NewSetup(a *SetupArgs) (*Setup, error) {
	if err := a.validate(); err != nil {
		return nil, log.Wrap(err)
	}
	awsClient, err := a.awsConnect()
	if err != nil {
		return nil, log.Wrap(err, "invalid AWS access credentials")
	}
	fs, err := domain.NewSingleDeveloperWorkspaceStore()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Setup{
		aws:         awsClient,
		accountName: a.AccountName,
		override:    a.Override,
		store:       fs,
	}, nil
}

func (c *Setup) Create(getPath func(string) (string, string)) error {
	ws := c.store.Workspace()
	bucket, key := getPath(c.aws.Region())
	ac, err := ws.NewAccount(c.accountName, c.aws.AccountID(), c.aws.Region(), bucket, key)
	if err != nil {
		return log.Wrap(err)
	}
	c.stackName = ac.SetupStackName()
	c.lambdaName = ac.SetupLambdaName()
	c.resourceTags = ac.ResourceTags()

	if err := c.create(ac); err != nil {
		return log.Wrap(err)
	}
	if err := c.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Setup) create(ac *domain.Account) error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	if exists {
		return log.Wrapf("Mantil is already installed in this AWS account")
	}
	if err := c.createSetupStack(ac.Functions); err != nil {
		return log.Wrap(err)
	}
	ui.Title("\nSetting up AWS infrastructure...\n")
	req := &dto.SetupRequest{
		Bucket:          ac.Bucket,
		FunctionsBucket: ac.Functions.Bucket,
		FunctionsPath:   ac.Functions.Path,
		AuthEnv:         ac.AuthEnv(),
		ResourceSuffix:  ac.ResourceSuffix(),
		ResourceTags:    c.resourceTags,
	}
	rsp := &dto.SetupResponse{}
	if err := backend.Lambda(c.aws.Lambda(), c.lambdaName).Call("create", req, rsp); err != nil {
		return log.Wrap(err, "failed to invoke setup function")
	}
	ac.Endpoints.Rest = rsp.APIGatewayRestURL
	ac.Endpoints.Ws = rsp.APIGatewayWsURL
	ac.CliRole = rsp.CliRole
	ui.Title("\nNode %s created with:", c.accountName)
	ui.Info(`
	+ Lambda functions
	+ API Gateways
	+ IAM Roles
	+ DynamoDB tables
	+ Cloudwatch log groups
	+ SQS forwarder
	+ S3 bucket
`)
	return nil
}

func (c *Setup) backendExists() (bool, error) {
	return c.aws.LambdaExists(c.lambdaName)
}

func (c *Setup) createSetupStack(acf domain.AccountFunctions) error {
	td := stackTemplateData{
		Name:   c.stackName,
		Bucket: acf.Bucket,
		S3Key:  fmt.Sprintf("%s/setup.zip", acf.Path),
		Region: c.aws.Region(),
	}
	t, err := c.renderStackTemplate(td)
	if err != nil {
		return log.Wrap(err, "render template failed")
	}
	stackWaiter := c.aws.CloudFormation().CreateStack(c.stackName, string(t), c.resourceTags)
	runStackProgress("Installing setup stack", stackWaiter)
	if err := stackWaiter.Wait(); err != nil {
		return log.Wrap(err, "cloudformation failed")
	}
	// https://github.com/aws-cloudformation/cloudformation-coverage-roadmap/issues/919
	if err := c.aws.TagLogGroup(aws.LambdaLogGroup(c.lambdaName), c.resourceTags); err != nil {
		return log.Wrap(err, "tagging setup lambda log group failed")
	}
	return nil
}

func (c *Setup) Destroy() error {
	ws := c.store.Workspace()
	ac := ws.Account(c.accountName)
	if ac == nil {
		return log.Wrapf("Account %s don't exists", c.accountName)
	}
	c.stackName = ac.SetupStackName()
	c.lambdaName = ac.SetupLambdaName()

	if err := c.destroy(ac); err != nil {
		return log.Wrap(err)
	}
	ws.RemoveAccount(ac.Name)
	if err := c.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Setup) destroy(ac *domain.Account) error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	if !exists {
		return log.Wrapf("Mantil not found in this AWS account")
	}

	req := &dto.SetupDestroyRequest{
		Bucket: ac.Bucket,
	}
	ui.Title("\nDestroying AWS infrastructure...\n")
	if err := backend.Lambda(c.aws.Lambda(), c.lambdaName).Call("destroy", req, nil); err != nil {
		return log.Wrap(err, "failed to call setup function")
	}
	stackWaiter := c.aws.CloudFormation().DeleteStack(c.stackName)
	runStackProgress("Destroying setup stack", stackWaiter)
	if err := stackWaiter.Wait(); err != nil {
		return log.Wrap(err)
	}
	ui.Notice("\n\nNode %s destroyed!", c.accountName)
	return nil
}

func (c *Setup) renderStackTemplate(data stackTemplateData) ([]byte, error) {
	return renderTemplate(setupStackTemplate, data)
}

type stackProgress struct {
	prefix      string
	currentCnt  uint32
	dotCnt      uint32
	stackWaiter *aws.StackWaiter
	done        chan struct{}
}

func runStackProgress(prefix string, stackWaiter *aws.StackWaiter) {
	sp := &stackProgress{
		prefix:      prefix,
		stackWaiter: stackWaiter,
		done:        make(chan struct{}),
	}
	sp.run()
}

func (p *stackProgress) run() {
	// hide cursor
	if runtime.GOOS != "windows" {
		fmt.Print("\033[?25l")
	}
	defer func() {
		fmt.Println()
		// show cursor
		if runtime.GOOS != "windows" {
			fmt.Print("\033[?25h")
		}
	}()
	fmt.Println()
	p.print()
	go p.printLoop()
	go p.handleStackEvents()
	<-p.done
	atomic.StoreUint32(&p.currentCnt, stackResourceCount)
	p.print()
}

func (p *stackProgress) printLoop() {
	ticker := time.NewTicker(time.Second)
	for {
		select {
		case <-ticker.C:
			atomic.StoreUint32(&p.dotCnt, (p.dotCnt+1)%4)
			p.print()
		case <-p.done:
			ticker.Stop()
			return
		}
	}
}

func (p *stackProgress) print() {
	var dots string
	format := "\r%s %d%% (%d/%d)%s"
	if p.currentCnt != stackResourceCount {
		dots = strings.Repeat(".", int(p.dotCnt))
		format = "\r%s %d%% (%d/%d)%-4s"
	}
	out := fmt.Sprintf(format,
		p.prefix,
		int(100*float64(p.currentCnt)/float64(stackResourceCount)),
		p.currentCnt,
		stackResourceCount,
		dots,
	)
	if p.currentCnt == stackResourceCount {
		out += ", done."
	}
	out = strings.ReplaceAll(out, "%", "%%")
	ui.Title(out)
}

func (p *stackProgress) handleStackEvents() {
	for range p.stackWaiter.Events() {
		atomic.AddUint32(&p.currentCnt, 1)
	}
	close(p.done)
}
