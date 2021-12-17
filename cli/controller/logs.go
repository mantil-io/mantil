package controller

import (
	"fmt"
	"time"

	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/cli/log"
)

type LogsArgs struct {
	Function string
	Filter   string
	Stage    string
	Tail     bool
	Since    time.Duration
}

type LogsCmd struct {
	awsClient *aws.AWS
	logGroup  string
	filter    string
	startTime time.Time
	tail      bool
}

func Logs(a LogsArgs) error {
	_, stage, err := newStoreWithStage(a.Stage)
	if err != nil {
		return log.Wrap(err)
	}
	awsClient, err := awsClient(stage.Node(), stage)
	if err != nil {
		return log.Wrap(err)
	}
	fn := stage.FindFunction(a.Function)
	if fn == nil {
		return log.Wrapf("function %s not found", a.Function)
	}
	logGroup := aws.LambdaLogGroup(fn.LambdaName())
	startTime := time.Now().Add(-a.Since)
	return printLogs(awsClient, logGroup, a.Filter, startTime, a.Tail)
}

func printLogs(awsClient *aws.AWS, logGroup, filter string, startTime time.Time, tail bool) error {
	startTs := startTime.UnixMilli()
	var lastEventTs int64

	fetchAndPrint := func(ts int64) error {
		events, err := awsClient.FetchLogs(logGroup, filter, &ts)
		if err != nil {
			return log.Wrap(err)
		}
		for e := range events {
			fmt.Printf("%v %v", time.UnixMilli(e.Timestamp), e.Message)
			lastEventTs = e.Timestamp
		}
		return nil
	}

	if err := fetchAndPrint(startTs); err != nil {
		return log.Wrap(err)
	}
	if !tail {
		return nil
	}
	for {
		startTs = lastEventTs + 1
		if err := fetchAndPrint(startTs); err != nil {
			return log.Wrap(err)
		}
		time.Sleep(time.Second)
	}
}
