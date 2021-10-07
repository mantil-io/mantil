package cmd

import (
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/workspace"
)

type logsFlags struct {
	function string
	filter   string
	since    time.Duration
	tail     bool
	stage    string
}

type logsCmd struct {
	ctx       *project.Context
	awsClient *aws.AWS
	function  string
	filter    string
	startTime time.Time
	tail      bool
}

func newLogs(f *logsFlags) (*logsCmd, error) {
	ctx, err := project.ContextWithStage(f.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	awsClient, err := ctx.AWSClient()
	if err != nil {
		return nil, log.Wrap(err)
	}
	if f.function == "" {
		f.function, err = selectFunctionFromStage(ctx.Stage)
		if err != nil {
			return nil, log.Wrap(err)
		}
	}
	return &logsCmd{
		ctx:       ctx,
		awsClient: awsClient,
		function:  f.function,
		filter:    f.filter,
		startTime: time.Now().Add(-f.since),
		tail:      f.tail,
	}, nil

}

func (c *logsCmd) run() error {
	startTs := c.timestamp(c.startTime)
	var lastEventTs int64

	fetchAndPrint := func(ts int64) error {
		events, err := c.awsClient.FetchLogs(c.logGroup(), c.filter, &ts)
		if err != nil {
			return err
		}
		for e := range events {
			c.printEvent(e)
			lastEventTs = e.Timestamp
		}
		return nil
	}

	if err := fetchAndPrint(startTs); err != nil {
		return err
	}
	if !c.tail {
		return nil
	}
	for {
		startTs = lastEventTs + 1
		if err := fetchAndPrint(startTs); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
}

func (c *logsCmd) timestamp(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

func (c *logsCmd) logGroup() string {
	lambdaName := workspace.ProjectResource(c.ctx.Project.Name, c.ctx.Stage.Name, c.function)
	return c.awsClient.LambdaLogGroup(lambdaName)
}

func (c *logsCmd) printEvent(e aws.LogEvent) {
	fmt.Printf(c.formatEvent(e))
}

func (c *logsCmd) formatEvent(e aws.LogEvent) string {
	return fmt.Sprintf("%v %v", c.eventTs(e), e.Message)
}

func (c *logsCmd) eventTs(e aws.LogEvent) time.Time {
	return time.Unix(0, e.Timestamp*int64(time.Millisecond))
}

func selectFunctionFromStage(stage *workspace.Stage) (string, error) {
	var funcNames []string
	for _, f := range stage.Functions {
		funcNames = append(funcNames, f.Name)
	}
	prompt := promptui.Select{
		Label: "Select a function",
		Items: funcNames,
	}
	_, function, err := prompt.Run()
	if err != nil {
		return "", log.Wrap(err)
	}
	return function, nil
}
