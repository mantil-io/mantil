package controller

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/controller/invoke"
	"github.com/mantil-io/mantil/cli/log"
	"github.com/mantil-io/mantil/cli/ui"
	"github.com/mantil-io/mantil/cli/ui/progress"
	"github.com/mantil-io/mantil/cli/ui/term"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

//go:embed setup_stack_template.yml
var setupStackTemplate string

const (
	stackResourceCount = 4
	APIGatewayLogsRole = "APIGatewayPushToCloudWatchLogsRole"
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
	Name               string
	Bucket             string
	S3Key              string
	Region             string
	Suffix             string
	APIGatewayLogsRole string
}

func NewSetup(a *SetupArgs) (*Setup, error) {
	if err := a.validate(); err != nil {
		return nil, log.Wrap(err)
	}
	awsClient, err := a.awsConnect()
	if err != nil {
		return nil, log.Wrap(err, "invalid AWS access credentials")
	}
	fs, err := newStore()
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
	n, err := ws.NewNode(c.nodeName, c.aws.AccountID(), c.aws.Region(), bucket, key)
	if err != nil {
		return log.Wrap(err)
	}
	c.stackName = n.SetupStackName()
	c.lambdaName = n.SetupLambdaName()
	c.resourceTags = n.ResourceTags()

	ui.Info("* If you need a break, take it now. This will take a minute, or two.")
	if err := c.create(n); err != nil {
		return log.Wrap(err)
	}
	if err := c.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Setup) create(n *domain.Node) error {
	tmr := timerFn()
	term.HideCursor()
	defer term.ShowCursor()
	if err := c.createSetupStack(n.Functions, n.ResourceSuffix()); err != nil {
		return log.Wrap(err)
	}
	stackDuration := tmr()

	ui.Info("")
	ui.Title("Setting up AWS infrastructure\n")
	req := &dto.SetupRequest{
		BucketConfig: dto.SetupBucketConfig{
			Name:         n.Bucket,
			ExpirePrefix: domain.FunctionsBucketPrefix,
			ExpireDays:   domain.FunctionsBucketExpireDays,
		},
		FunctionsBucket:    n.Functions.Bucket,
		FunctionsPath:      n.Functions.Path,
		AuthEnv:            n.AuthEnv(),
		ResourceSuffix:     n.ResourceSuffix(),
		APIGatewayLogsRole: APIGatewayLogsRole,
		ResourceTags:       c.resourceTags,
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

	ui.Info("")
	ui.Title("Mantil node %s created:", c.nodeName)
	c.printNodeResources()
	return nil
}

func (c *Setup) backendExists() (bool, error) {
	return c.aws.LambdaExists(c.lambdaName)
}

func (c *Setup) createSetupStack(acf domain.NodeFunctions, suffix string) error {
	td := stackTemplateData{
		Name:               c.stackName,
		Bucket:             acf.Bucket,
		S3Key:              fmt.Sprintf("%s/setup.zip", acf.Path),
		Region:             c.aws.Region(),
		Suffix:             suffix,
		APIGatewayLogsRole: APIGatewayLogsRole,
	}
	t, err := c.renderStackTemplate(td)
	if err != nil {
		return log.Wrap(err, "render template failed")
	}
	stackWaiter := c.aws.CloudFormation().CreateStack(c.stackName, string(t), c.resourceTags)
	runStackProgress("Installing setup stack:", stackWaiter)
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
	term.HideCursor()
	defer term.ShowCursor()
	ws := c.store.Workspace()
	if len(ws.Nodes) == 0 {
		return log.Wrapf("Nothing to delete, there are no nodes installed in your workspace")
	}
	n := ws.Node(c.nodeName)
	if n == nil {
		return log.Wrapf("Node %s doesn't exist. For a complete list of available nodes run 'mantil aws ls'", c.nodeName)
	}
	ok, err := c.confirmDestroy(n)
	if err != nil {
		return log.Wrap(err)
	}
	if !ok {
		return nil
	}
	c.stackName = n.SetupStackName()
	c.lambdaName = n.SetupLambdaName()

	ui.Info("* Grab your seat and stay patient. This will take a while.")
	if err := c.destroy(n); err != nil {
		return log.Wrap(err)
	}
	ws.RemoveNode(n.Name)
	if err := c.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Setup) confirmDestroy(n *domain.Node) (bool, error) {
	ui.Info("? You are going to destroy node %s. This action cannot be reversed.", n.Name)
	if len(n.Stages) != 0 {
		ui.Info("This node contains deployed stages which will be orphaned if the node is destroyed.")
	}
	confirmationPrompt := promptui.Prompt{
		Label: "To confirm, type 'yes'",
	}
	res, err := confirmationPrompt.Run()
	if err != nil {
		return false, log.Wrap(err)
	}
	res = strings.ToLower(res)
	if res != "yes" && res != "y" {
		return false, nil
	}
	return true, nil
}

func (c *Setup) selectNodeForDestroy(text string, nodes []string) (string, error) {
	prompt := promptui.Select{
		Label: text,
		Items: nodes,
	}
	_, node, err := prompt.Run()
	if err != nil {
		return "", log.Wrap(err)
	}
	return node, nil
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
	runStackProgress("Destroying setup stack:", stackWaiter)
	if err := stackWaiter.Wait(); err != nil {
		return log.Wrap(err)
	}
	stackDuration := tmr()

	log.Event(domain.Event{NodeDelete: &domain.NodeEvent{
		AWSCredentialsProvider: c.credentialsProvider,
		StackDuration:          stackDuration,
		InfrastructureDuration: infrastructureDuration,
	}})
	ui.Info("")
	ui.Title("Mantil node %s destroyed:", c.nodeName)
	c.printNodeResources()
	return nil
}

func (c *Setup) renderStackTemplate(data stackTemplateData) ([]byte, error) {
	return renderTemplate(setupStackTemplate, data)
}

func (c *Setup) printNodeResources() {
	ui.Info(`
	+ S3 bucket
	+ Lambda functions
	+ API Gateways
	+ IAM Roles
	+ DynamoDB tables
	+ Cloudwatch log groups
	+ Cloudformation stack`)
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
