package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

func (a *AWS) FetchLogs(group string, filter string, start *int64) (chan types.FilteredLogEvent, error) {
	streams, err := a.listLogStreams(group, start)
	if err != nil {
		return nil, err
	}
	events := make(chan types.FilteredLogEvent)
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
			events <- e
		}
		close(events)
	}()
	return events, nil
}

func (a *AWS) listLogStreams(group string, start *int64) ([]types.LogStream, error) {
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
