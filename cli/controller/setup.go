package controller

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/cli/ui/progress"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

//go:embed setup_stack_template.yml
var setupStackTemplate string

const (
	stackResourceCount = 4
)

type Setup struct {
	aws                 *aws.AWS
	nodeName            string
	override            bool // TODO unused
	store               *domain.FileStore
	resourceTags        map[string]string
	stackName           string
	lambdaName          string
	credentialsProvider int
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
		aws:                 awsClient,
		nodeName:            a.NodeName,
		override:            a.Override,
		store:               fs,
		credentialsProvider: a.credentialsProvider,
	}, nil
}

func (c *Setup) Create(getPath func(string) (string, string)) error {
	ws := c.store.Workspace()
	bucket, key := getPath(c.aws.Region())
	ac, err := ws.NewNode(c.nodeName, c.aws.AccountID(), c.aws.Region(), bucket, key)
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

func (c *Setup) create(n *domain.Node) error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	if exists {
		return log.Wrapf("Mantil is already installed in this AWS account")
	}

	tmr := timerFn()
	ui.HideCursor()
	defer ui.ShowCursor()
	if err := c.createSetupStack(n.Functions); err != nil {
		return log.Wrap(err)
	}
	stackDuration := tmr()

	ui.Title("Setting up AWS infrastructure\n")
	req := &dto.SetupRequest{
		BucketConfig: dto.SetupBucketConfig{
			Name:         n.Bucket,
			ExpirePrefix: domain.FunctionsBucketPrefix,
			ExpireDays:   domain.FunctionsBucketExpireDays,
		},
		FunctionsBucket: n.Functions.Bucket,
		FunctionsPath:   n.Functions.Path,
		AuthEnv:         n.AuthEnv(),
		ResourceSuffix:  n.ResourceSuffix(),
		ResourceTags:    c.resourceTags,
	}
	rsp := &dto.SetupResponse{}
	if err := invoke.Lambda(c.aws.Lambda(), c.lambdaName, ui.NodeLogsSink).Do("create", req, rsp); err != nil {
		return log.Wrap(err, "failed to invoke setup function")
	}
	n.Endpoints.Rest = rsp.APIGatewayRestURL
	n.CliRole = rsp.CliRole
	infrastructureDuration := tmr()

	log.Event(domain.Event{NodeCreate: &domain.NodeEvent{
		AWSCredentialsProvider: c.credentialsProvider,
		StackDuration:          stackDuration,
		InfrastructureDuration: infrastructureDuration,
		AWSRegion:              c.aws.Region(),
	}})

	ui.Title("\nNode %s created with:", c.nodeName)
	ui.Info(`
	+ Lambda functions
	+ API Gateways
	+ IAM Roles
	+ DynamoDB tables
	+ Cloudwatch log groups
	+ S3 bucket
`)
	return nil
}

func (c *Setup) backendExists() (bool, error) {
	return c.aws.LambdaExists(c.lambdaName)
}

func (c *Setup) createSetupStack(acf domain.NodeFunctions) error {
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
	ui.HideCursor()
	defer ui.ShowCursor()
	ws := c.store.Workspace()
	n := ws.Node(c.nodeName)
	if n == nil {
		return log.Wrapf("Node %s don't exists", c.nodeName)
	}
	c.stackName = n.SetupStackName()
	c.lambdaName = n.SetupLambdaName()

	if err := c.destroy(n); err != nil {
		return log.Wrap(err)
	}
	ws.RemoveNode(n.Name)
	if err := c.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Setup) destroy(n *domain.Node) error {
	exists, err := c.backendExists()
	if err != nil {
		return log.Wrap(err)
	}
	if !exists {
		return log.Wrapf("Mantil not found in this AWS account")
	}

	tmr := timerFn()
	req := &dto.SetupDestroyRequest{
		Bucket: n.Bucket,
	}
	ui.Title("\nDestroying AWS infrastructure\n")
	if err := invoke.Lambda(c.aws.Lambda(), c.lambdaName, ui.NodeLogsSink).Do("destroy", req, nil); err != nil {
		return log.Wrap(err, "failed to call setup function")
	}
	infrastructureDuration := tmr()

	stackWaiter := c.aws.CloudFormation().DeleteStack(c.stackName)
	runStackProgress("Destroying setup stack", stackWaiter)
	if err := stackWaiter.Wait(); err != nil {
		return log.Wrap(err)
	}
	stackDuration := tmr()

	log.Event(domain.Event{NodeDelete: &domain.NodeEvent{
		AWSCredentialsProvider: c.credentialsProvider,
		StackDuration:          stackDuration,
		InfrastructureDuration: infrastructureDuration,
	}})

	ui.Notice("\nNode %s destroyed!", c.nodeName)
	return nil
}

func (c *Setup) renderStackTemplate(data stackTemplateData) ([]byte, error) {
	return renderTemplate(setupStackTemplate, data)
}

type stackProgress struct {
	prefix      string
	currentCnt  int
	stackWaiter *aws.StackWaiter
	counter     *progress.Counter
	progress    *progress.Progress
	lines       chan string
}

func runStackProgress(prefix string, stackWaiter *aws.StackWaiter) {
	sp := &stackProgress{
		prefix:      prefix,
		stackWaiter: stackWaiter,
		lines:       make(chan string),
	}
	sp.counter = progress.NewCounter(stackResourceCount)
	sp.progress = progress.New(prefix, progress.LogFuncBold(), sp.counter, progress.NewDots())
	sp.run()
}

func (p *stackProgress) run() {
	log.Printf(p.prefix)
	fmt.Println()
	p.progress.Run()
	p.handleStackEvents()
	p.progress.Stop()
	fmt.Println()
	log.Printf("%s: done", p.prefix)
}

func (p *stackProgress) handleStackEvents() {
	for e := range p.stackWaiter.Events() {
		ebuf, _ := json.Marshal(e)
		log.Printf("%s event: %s", p.prefix, string(ebuf))
		if e.ResourceStatus != types.ResourceStatusCreateComplete &&
			e.ResourceStatus != types.ResourceStatusDeleteComplete {
			continue
		}
		if p.currentCnt < stackResourceCount {
			p.currentCnt++
			p.counter.SetCount(p.currentCnt)
		}
	}
	p.currentCnt = stackResourceCount
	p.counter.SetCount(p.currentCnt)
	close(p.lines)
}
