package logs

import (
	"fmt"
	"time"

	"github.com/araddon/dateparse"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/mantil-io/mantil/internal/aws"
)

type LogsCmd struct {
	awsClient *aws.AWS
}

func New(awsClient *aws.AWS) *LogsCmd {
	return &LogsCmd{
		awsClient: awsClient,
	}
}

func (c *LogsCmd) Fetch(function, filter string, since time.Duration, tail bool) error {
	group := fmt.Sprintf("/aws/lambda/mantil-project-%s", function)
	st := time.Now().Add(-since).UnixNano() / int64(time.Millisecond)
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
