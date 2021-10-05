package cmd

import (
	"fmt"
	"github.com/mantil-io/mantil/cli/cmd/project"
	"time"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/workspace"
)

type logsCmd struct {
	ctx       *project.Context
	function  string
	awsClient *aws.AWS
	filter    string
	startTime time.Time
	tail      bool
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
