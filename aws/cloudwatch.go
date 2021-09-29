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
	streams, err := a.listLogStreams(group)
	if err != nil {
		return nil, err
	}
	events := make(chan LogEvent)
	go func() {
		var streamNames []string
		for _, s := range streams {
			streamNames = append(streamNames, *s.LogStreamName)
		}
		es, err := a.fetchLogStreams(group, filter, streamNames, start)
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

func (a *AWS) listLogStreams(group string) ([]types.LogStream, error) {
	var streams []types.LogStream
	var nextToken *string
	for {
		dlsi := &cloudwatchlogs.DescribeLogStreamsInput{
			LogGroupName: aws.String(group),
			Descending:   aws.Bool(true),
			NextToken:    nextToken,
			OrderBy:      types.OrderByLastEventTime,
		}
		out, err := a.cloudwatchClient.DescribeLogStreams(context.Background(), dlsi)
		if err != nil {
			return nil, err
		}
		for _, s := range out.LogStreams {
			streams = append([]types.LogStream{s}, streams...)
		}
		if out.NextToken == nil {
			break
		}
		nextToken = out.NextToken
	}
	return streams, nil
}

func (a *AWS) fetchLogStreams(group, filter string, streams []string, start *int64) ([]types.FilteredLogEvent, error) {
	var events []types.FilteredLogEvent
	var nextToken *string
	for {
		flei := &cloudwatchlogs.FilterLogEventsInput{
			FilterPattern:  aws.String(filter),
			LogGroupName:   aws.String(group),
			LogStreamNames: streams,
			NextToken:      nextToken,
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
