package aws

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
)

func (a *AWS) LambdaLogGroup(lambdaName string) string {
	return fmt.Sprintf("/aws/lambda/%s", lambdaName)
}

func (a *AWS) FetchLogs(group string, filter string, start int64, tail bool) (chan LogEvent, error) {
	if tail {
		start = time.Now().UnixMilli()
	}
	lef := a.logEventsFunc(group, filter, start)
	events := make(chan LogEvent)

	fetchAndSend := func() {
		for e := range lef() {
			le := LogEvent{
				Message:   aws.ToString(e.Message),
				Timestamp: aws.ToInt64(e.Timestamp),
			}
			events <- le
		}
	}

	go func() {
		fetchAndSend()

		if !tail {
			close(events)
			return
		}
		time.Sleep(1 * time.Second)

		for {
			fetchAndSend()
			time.Sleep(1 * time.Second)
		}
	}()
	return events, nil
}

// since each lambda invocation writes logs to its own log stream with a small delay situations in which some messages get skipped can arise
// for that reason we query next batch not from the max ts of all events, but minimum ts of last events among all streams
// that way we include those events which might have been skipped but also get some repeated which then must be manually skipped
// new streams might still have chunk of their messages skipped if they appear between the requests, but that's cloudwatch limitation
func (a *AWS) logEventsFunc(group string, filter string, start int64) func() chan types.FilteredLogEvent {
	startTs := start
	streamsTs := make(map[string]int64, 0)
	seenEventIDs := make(map[string]struct{})

	return func() chan types.FilteredLogEvent {
		events := make(chan types.FilteredLogEvent)
		go func() {
			eventIDs := make(map[string]struct{})
			es, err := a.fetchLogEvents(group, filter, &startTs)
			if err != nil {
				close(events)
				return
			}
			for _, e := range es {
				streamsTs[aws.ToString(e.LogStreamName)] = aws.ToInt64(e.Timestamp)

				eventID := aws.ToString(e.EventId)
				eventIDs[eventID] = struct{}{}
				if _, seen := seenEventIDs[eventID]; seen {
					continue
				}

				events <- e
			}
			seenEventIDs = eventIDs
			if len(streamsTs) > 0 {
				startTs = earliestLogStreamTs(streamsTs) + 1
			} else {
				startTs += 1
			}
			streamsTs = make(map[string]int64, 0)
			close(events)
		}()
		return events
	}
}

func (a *AWS) fetchLogEvents(group, filter string, start *int64) ([]types.FilteredLogEvent, error) {
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

func earliestLogStreamTs(m map[string]int64) int64 {
	var ets int64 = math.MaxInt64
	for _, ts := range m {
		if ts < ets {
			ets = ts
		}
	}
	return ets
}

type LogEvent struct {
	Message   string
	Timestamp int64
}
