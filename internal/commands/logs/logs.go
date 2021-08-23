package logs

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/mantil-io/mantil-cli/internal/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

type LogsCmd struct {
	awsClient *aws.AWS
}

func New(awsClient *aws.AWS) *LogsCmd {
	return &LogsCmd{
		awsClient: awsClient,
	}
}

func (c *LogsCmd) Fetch(function, filter, start string, tail bool) error {
	t, err := c.parseTime(start)
	if err != nil {
		return fmt.Errorf("could not parse start time - %v", err)
	}
	group := fmt.Sprintf("/aws/lambda/%s", function)
	st := t.UnixNano() / int64(time.Millisecond)
	var lastEventTs int64
	fetchAndPrint := func(ts *int64) error {
		events, err := c.awsClient.FetchLogs(group, filter, ts)
		if err != nil {
			return err
		}
		for e := range events {
			printEvent(e)
			lastEventTs = *e.Timestamp + 1
		}
		return nil
	}
	if err := fetchAndPrint(&st); err != nil {
		return err
	}
	if !tail {
		return nil
	}
	for {
		if err := fetchAndPrint(&lastEventTs); err != nil {
			return err
		}
		time.Sleep(time.Second)
	}
}

func (c *LogsCmd) parseTime(timeStr string) (*time.Time, error) {
	if timeStr == "" {
		t := time.Now().Add(-3 * time.Hour)
		return &t, nil
	}
	t, err := dateparse.ParseLocal(timeStr)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func printEvent(e types.FilteredLogEvent) {
	fmt.Printf("%v %v", eventTs(e), *e.Message)
}

func eventTs(e types.FilteredLogEvent) *time.Time {
	t := time.Unix(0, *e.Timestamp*int64(time.Millisecond))
	return &t
}
