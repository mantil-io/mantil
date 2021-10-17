package cmd

import (
	"fmt"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"github.com/mantil-io/mantil/cli/log"

	"github.com/mantil-io/mantil/aws"
)

type logsArgs struct {
	function string
	filter   string
	since    time.Duration
	tail     bool
	stage    string
}

type logsCmd struct {
	awsClient *aws.AWS
	logGroup  string
	function  string
	filter    string
	startTime time.Time
	tail      bool
}

func newLogs(a logsArgs) (*logsCmd, error) {
	fs, err := project.NewStoreWithStage(a.stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	stage := fs.Stage(a.stage)
	awsClient, err := project.AWSClient(stage.Account(), stage.Project(), stage)
	if err != nil {
		return nil, log.Wrap(err)
	}
	if a.function == "" {
		names := stage.FunctionNames()
		if len(names) == 1 {
			a.function = names[0]
		} else {
			a.function, err = selectFunctionFromStage(names)
			if err != nil {
				return nil, log.Wrap(err)
			}
		}
	}

	fn := stage.FindFunction(a.function)
	if fn == nil {
		return nil, log.WithUserMessage(nil, "function %s not found", a.function)
	}

	return &logsCmd{
		awsClient: awsClient,
		function:  a.function,
		filter:    a.filter,
		startTime: time.Now().Add(-a.since),
		tail:      a.tail,
		logGroup:  aws.LambdaLogGroup(fn.LambdaName()),
	}, nil

}

func (c *logsCmd) run() error {
	startTs := c.timestamp(c.startTime)
	var lastEventTs int64

	fetchAndPrint := func(ts int64) error {
		events, err := c.awsClient.FetchLogs(c.logGroup, c.filter, &ts)
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

func (c *logsCmd) printEvent(e aws.LogEvent) {
	fmt.Printf(c.formatEvent(e))
}

func (c *logsCmd) formatEvent(e aws.LogEvent) string {
	return fmt.Sprintf("%v %v", c.eventTs(e), e.Message)
}

func (c *logsCmd) eventTs(e aws.LogEvent) time.Time {
	return time.Unix(0, e.Timestamp*int64(time.Millisecond))
}

func selectFunctionFromStage(funcNames []string) (string, error) {
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
