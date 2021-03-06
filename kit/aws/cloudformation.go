package aws

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
)

type CloudFormation struct {
	aws *AWS
	cli *cloudformation.Client
}

func (a *AWS) CloudFormation() *CloudFormation {
	return &CloudFormation{
		aws: a,
		cli: a.cloudformationClient,
	}
}

func (f *CloudFormation) CreateStack(name, templateBody string, resourceTags map[string]string) *StackWaiter {
	csi := &cloudformation.CreateStackInput{
		StackName:    aws.String(name),
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
		OnFailure:    types.OnFailureDelete,
		TemplateBody: aws.String(templateBody),
	}
	if resourceTags != nil {
		tags := []types.Tag{}
		for k, v := range resourceTags {
			tags = append(tags, types.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		csi.Tags = tags
	}

	sw := newStackWaiter(name, f)
	go func() {
		cso, err := f.cli.CreateStack(context.Background(), csi)
		if err != nil {
			sw.close(err)
			return
		}
		w := cloudformation.NewStackCreateCompleteWaiter(f.cli, func(opts *cloudformation.StackCreateCompleteWaiterOptions) {
			opts.MinDelay = 10 * time.Second
			opts.MaxDelay = 20 * time.Second
		})
		dsi := &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		}
		if err := w.Wait(context.Background(), dsi, 5*time.Minute); err != nil {
			reason := f.stackActionFailedReason(aws.ToString(cso.StackId))
			if reason != "" {
				sw.close(fmt.Errorf("could not create stack %s - %s", name, reason))
				return
			}
			sw.close(fmt.Errorf("could not create stack %s", name))
			return
		}
		sw.close(nil)
	}()
	return sw
}

func (f *CloudFormation) UpdateStack(name, templateBody string, resourceTags map[string]string) (*StackWaiter, error) {
	usi := &cloudformation.UpdateStackInput{
		StackName:    aws.String(name),
		Capabilities: []types.Capability{types.CapabilityCapabilityNamedIam},
		TemplateBody: aws.String(templateBody),
	}
	if resourceTags != nil {
		tags := []types.Tag{}
		for k, v := range resourceTags {
			tags = append(tags, types.Tag{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		usi.Tags = tags
	}

	sw := newStackWaiter(name, f)
	go func() {
		cso, err := f.cli.UpdateStack(context.Background(), usi)
		if err != nil {
			sw.close(err)
			return
		}
		w := cloudformation.NewStackUpdateCompleteWaiter(f.cli, func(opts *cloudformation.StackUpdateCompleteWaiterOptions) {
			opts.MinDelay = 10 * time.Second
			opts.MaxDelay = 20 * time.Second
		})
		dsi := &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		}
		if err := w.Wait(context.Background(), dsi, 5*time.Minute); err != nil {
			reason := f.stackActionFailedReason(aws.ToString(cso.StackId))
			if reason != "" {
				sw.close(fmt.Errorf("could not update stack %s - %s", name, reason))
				return
			}
			sw.close(fmt.Errorf("could not update stack %s", name))
			return
		}
		sw.close(nil)
	}()
	return sw, nil
}

func (f *CloudFormation) DeleteStack(name string) *StackWaiter {
	sw := newStackWaiter(name, f)
	go func() {
		dsi := &cloudformation.DeleteStackInput{
			StackName: aws.String(name),
		}
		_, err := f.cli.DeleteStack(context.Background(), dsi)
		if err != nil {
			sw.close(fmt.Errorf("could not delete stack %s, error: %w", name, err))
			return
		}
		w := cloudformation.NewStackDeleteCompleteWaiter(f.cli, func(opts *cloudformation.StackDeleteCompleteWaiterOptions) {
			opts.MinDelay = 10 * time.Second
			opts.MaxDelay = 20 * time.Second
		})
		descsi := &cloudformation.DescribeStacksInput{
			StackName: aws.String(name),
		}
		if err := w.Wait(context.Background(), descsi, 5*time.Minute); err != nil {
			reason := f.stackActionFailedReason(name)
			if reason != "" {
				sw.close(fmt.Errorf("could not delete stack %s - %s", name, reason))
				return
			}
			sw.close(fmt.Errorf("could not delete stack %s", name))
			return
		}
		sw.close(nil)
	}()
	return sw
}

// tries to find reason why stack action failed by going through all the events until first one with status failed is encountered
// asumption is made that first failed event is also the reason for the failure of the stack action
// for stacks that do not longer exists (like in the case of create failure) stack id must be provided instead of name
func (f *CloudFormation) stackActionFailedReason(stack string) string {
	es, err := f.stackEvents(stack, nil)
	if err != nil {
		log.Printf("DescribeStackEvents error: %v", err)
		return ""
	}
	if len(es) == 0 {
		return ""
	}
	for i := len(es) - 1; i >= 0; i-- {
		event := es[i]
		if strings.Contains(string(event.ResourceStatus), "FAILED") {
			return fmt.Sprintf("%s - %s",
				aws.ToString(event.LogicalResourceId),
				aws.ToString(event.ResourceStatusReason))
		}
	}
	return ""
}

func (f *CloudFormation) stackEvents(stack string, from *time.Time) ([]types.StackEvent, error) {
	var events []types.StackEvent
	var nextToken *string
	for {
		dsei := &cloudformation.DescribeStackEventsInput{
			StackName: aws.String(stack),
			NextToken: nextToken,
		}
		dseo, err := f.cli.DescribeStackEvents(context.Background(), dsei)
		if err != nil {
			return nil, err
		}
		events = append(events, dseo.StackEvents...)
		if dseo.NextToken == nil {
			break
		}
		nextToken = dseo.NextToken
	}
	if from != nil {
		for idx, e := range events {
			if e.Timestamp.Before(*from) || e.Timestamp.Equal(*from) {
				return events[:idx], nil
			}
		}
	}
	return events, nil
}

type StackWaiter struct {
	stackName      string
	cloudFormation *CloudFormation
	errc           chan error
	done           chan struct{}
	loopDone       chan struct{}
	events         chan types.StackEvent
	closer         sync.Once
}

func newStackWaiter(stackName string, cf *CloudFormation) *StackWaiter {
	sw := &StackWaiter{
		stackName:      stackName,
		cloudFormation: cf,
		errc:           make(chan error),
		done:           make(chan struct{}),
		loopDone:       make(chan struct{}),
		events:         make(chan types.StackEvent),
	}
	go sw.pollEvents()
	return sw
}

func (w *StackWaiter) Wait() error {
	return <-w.errc
}

func (w *StackWaiter) close(err error) {
	w.closer.Do(func() {
		close(w.done)
		<-w.loopDone
		w.errc <- err
	})
}

func (w *StackWaiter) Events() <-chan types.StackEvent {
	return w.events
}

func (w *StackWaiter) pollEvents() {
	ticker := time.NewTicker(time.Second)
	ts := time.Now()
	poll := func() {
		es, _ := w.cloudFormation.stackEvents(w.stackName, &ts)
		for _, e := range es {
			w.events <- e
		}
		if len(es) > 0 {
			ts = *es[0].Timestamp
		}
	}
	for {
		select {
		case <-ticker.C:
			poll()
		case <-w.done:
			poll()
			ticker.Stop()
			close(w.events)
			close(w.loopDone)
			return
		}
	}
}
