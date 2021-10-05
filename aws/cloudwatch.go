package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

func (a *AWS) LambdaLogGroup(lambdaName string) string {
	return fmt.Sprintf("/aws/lambda/%s", lambdaName)
}

func (a *AWS) FetchLogs(group string, filter string, start *int64) (chan LogEvent, error) {
	events := make(chan LogEvent)
	go func() {
		es, err := a.fetchLogStreams(group, filter, start)
		if err != nil {
			log.Println(err)
			return
		}
		for _, e := range es {
			le := LogEvent{
				Message:   aws.ToString(e.Message),
				Timestamp: aws.ToInt64(e.Timestamp),
			}
			events <- le
		}
		close(events)
	}()
	return events, nil
}

func (a *AWS) fetchLogStreams(group, filter string, start *int64) ([]types.FilteredLogEvent, error) {
	var events []types.FilteredLogEvent
	var nextToken *string
	for {
		flei := &cloudwatchlogs.FilterLogEventsInput{
			LogGroupName:  aws.String(group),
			FilterPattern: aws.String(filter),
			NextToken:     nextToken,
		}
		if filter != "" {
			flei.FilterPattern = aws.String(filter)
		}
		if start != nil {
			flei.StartTime = start
		}
		out, err := a.cloudwatchClient.FilterLogEvents(context.Background(), flei)
		if err != nil {
			return nil, err
		}
		events = append(events, out.Events...)
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return events, nil
}

type LogEvent struct {
	Message   string
	Timestamp int64
}
