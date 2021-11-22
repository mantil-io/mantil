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
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/node/dto"
)

//go:embed setup_stack_template.yml
var setupStackTemplate string

const (
	stackResourceCount = 4
	APIGatewayLogsRole = "APIGatewayPushToCloudWatchLogsRole"
)

var (
	supportedAWSRegions = []string{"ap-south-1", "ap-southeast-1", "ap-southeast-2", "ap-northeast-1", "eu-central-1", "eu-west-1", "eu-west-2", "us-east-1", "us-east-2", "us-west-2"}
)

type Setup struct {
	aws                 *aws.AWS
	nodeName            string
	store               *domain.FileStore
	resourceTags        map[string]string
	stackName           string
	lambdaName          string
	credentialsProvider int
	force               bool
	yes                 bool
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
		store:               fs,
		credentialsProvider: a.credentialsProvider,
		force:               a.Force,
		yes:                 a.Yes,
	}, nil
}

func (c *Setup) Create(getPath func(string) (string, string)) error {
	if !c.regionSupported() {
		return log.Wrapf(`Mantil is currently not available in this region.
Available regions are:
	+ %s`, strings.Join(supportedAWSRegions, "\n\t+ "))
	}
	ws := c.store.Workspace()
	bucket, key := getPath(c.aws.Region())
	n, err := ws.NewNode(c.nodeName, c.aws.AccountID(), c.aws.Region(), bucket, key)
	if err != nil {
		return log.Wrap(err)
	}
	c.stackName = n.SetupStackName()
	c.lambdaName = n.SetupLambdaName()
	c.resourceTags = n.ResourceTags()

	if err := c.create(n); err != nil {
		return log.Wrap(err)
	}
	if err := c.store.Store(); err != nil {
		return log.Wrap(err)
	}
	return nil
}

func (c *Setup) regionSupported() bool {
	region := c.aws.Region()
	for _, r := range supportedAWSRegions {
		if region == r {
			return true
		}
	}
	return false
}

func (c *Setup) create(n *domain.Node) error {
	tmr := timerFn()
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
	ui.Title("Mantil node %s created:\n", c.nodeName)
	c.printNodeResources("+")
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
	if err := runStackProgress("Installing setup stack", types.ResourceStatusCreateComplete, stackWaiter); err != nil {
		log.Error(err)
		return log.Wrapf("Installing setup stack failed. No resources were created in your AWS account.")
	}
	// https://github.com/aws-cloudformation/cloudformation-coverage-roadmap/issues/919
	if err := c.aws.TagLogGroup(aws.LambdaLogGroup(c.lambdaName), c.resourceTags); err != nil {
		return log.Wrap(err, "tagging setup lambda log group failed")
	}
	return nil
}

func (c *Setup) Destroy() (bool, error) {
	ws := c.store.Workspace()
	if len(ws.Nodes) == 0 {
		return false, log.Wrapf("Nothing to delete, there are no nodes installed in your workspace")
	}
	n := ws.Node(c.nodeName)
	if n == nil {
		return false, log.Wrapf("Node %s doesn't exist. For a complete list of available nodes run 'mantil aws ls'", c.nodeName)
	}
	if !c.confirmDestroy(n) {
		return false, nil
	}
	c.stackName = n.SetupStackName()
	c.lambdaName = n.SetupLambdaName()

	if err := c.destroy(n); err != nil {
		return false, log.Wrap(err)
	}
	ws.RemoveNode(n.Name)
	if err := c.store.Store(); err != nil {
		return false, log.Wrap(err)
	}
	return true, nil
}

func (c *Setup) confirmDestroy(n *domain.Node) bool {
	if len(n.Stages) == 0 && (c.yes || c.force) {
		return true
	}
	if len(n.Stages) != 0 {
		if c.force {
			return true
		} else {
			ui.Info("This node contains deployed stages which will be orphaned if the node is destroyed.")
			ui.Info("As a measure of precaution please use option '--force' for this action.")
			return false
		}
	}
	ui.Title("? Do you really want to destroy node %s?\n", n.Name)
	ui.Info("This action cannot be reversed.")
	confirmationPrompt := promptui.Prompt{
		Label: "To confirm, type 'yes'",
	}
	res, err := confirmationPrompt.Run()
	if err != nil {
		return false
	}
	res = strings.ToLower(res)
	if res != "yes" && res != "y" {
		return false
	}
	return true
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
	if err := runStackProgress("Destroying setup stack", types.ResourceStatusDeleteComplete, stackWaiter); err != nil {
		return log.Wrap(err)
	}
	stackDuration := tmr()

	log.Event(domain.Event{NodeDelete: &domain.NodeEvent{
		AWSCredentialsProvider: c.credentialsProvider,
		StackDuration:          stackDuration,
		InfrastructureDuration: infrastructureDuration,
	}})
	ui.Info("")
	ui.Title("Mantil node %s destroyed:\n", c.nodeName)
	c.printNodeResources("-")
	return nil
}

func (c *Setup) renderStackTemplate(data stackTemplateData) ([]byte, error) {
	return renderTemplate(setupStackTemplate, data)
}

func (c *Setup) printNodeResources(sign string) {
	resources := []string{
		"S3 bucket",
		"Lambda functions",
		"API Gateways",
		"IAM Roles",
		"DynamoDB tables",
		"Cloudwatch log groups",
		"Cloudformation stack",
	}
	sep := fmt.Sprintf("\n\t%s ", sign)
	ui.Info("\t%s %s ", sign, strings.Join(resources, sep))
}

type stackProgress struct {
	prefix      string
	currentCnt  int
	stackWaiter *aws.StackWaiter
	counter     *progress.Counter
	progress    *progress.Progress
	status      types.ResourceStatus
	lines       chan string
}

func runStackProgress(prefix string, status types.ResourceStatus, stackWaiter *aws.StackWaiter) error {
	sp := &stackProgress{
		prefix:      prefix,
		stackWaiter: stackWaiter,
		status:      status,
		lines:       make(chan string),
	}
	if !strings.HasSuffix(prefix, ":") {
		prefix = fmt.Sprintf("%s:", prefix)
	}
	sp.counter = progress.NewCounter(stackResourceCount)
	sp.progress = progress.New(prefix, progress.LogFuncBold(), sp.counter, progress.NewDots())
	return sp.run()
}

func (p *stackProgress) run() error {
	log.Printf(p.prefix)
	fmt.Println()
	p.progress.Run()
	p.handleStackEvents()
	err := p.stackWaiter.Wait()
	if err == nil {
		p.currentCnt = stackResourceCount
		p.counter.SetCount(p.currentCnt)
		p.progress.Done()
	} else {
		p.progress.Abort()
	}
	log.Printf("%s: done", p.prefix)
	return err
}

func (p *stackProgress) handleStackEvents() {
	for e := range p.stackWaiter.Events() {
		ebuf, _ := json.Marshal(e)
		log.Printf("%s event: %s", p.prefix, string(ebuf))
		if e.ResourceStatus != p.status {
			continue
		}
		if p.currentCnt < stackResourceCount {
			p.currentCnt++
			p.counter.SetCount(p.currentCnt)
		}
	}
	close(p.lines)
}
