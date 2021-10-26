package aws

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation"
	"github.com/aws/aws-sdk-go-v2/service/cloudformation/types"
	"github.com/mantil-io/mantil/cli/log"
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

func (f *CloudFormation) CreateStack(name, templateBody string, resourceTags map[string]string) error {
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
	cancel := f.pollStackEvents(name)
	defer cancel()
	cso, err := f.cli.CreateStack(context.Background(), csi)
	if err != nil {
		return log.Wrap(err, fmt.Sprintf("could not create stack %s", name))
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
			return log.Wrap(fmt.Errorf("could not create stack %s - %s", name, reason))
		}
		return log.Wrap(fmt.Errorf("could not create stack %s", name))
	}
	return nil
}

func (f *CloudFormation) DeleteStack(name string) error {
	dsi := &cloudformation.DeleteStackInput{
		StackName: aws.String(name),
	}
	cancel := f.pollStackEvents(name)
	defer cancel()
	_, err := f.cli.DeleteStack(context.Background(), dsi)
	if err != nil {
		return log.Wrap(err, fmt.Sprintf("could not delete stack %s", name))
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
			return log.Wrap(fmt.Errorf("could not delete stack %s - %s", name, reason))
		}
		return log.Wrap(fmt.Errorf("could not delete stack %s", name))
	}
	return nil
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

func (f *CloudFormation) pollStackEvents(stack string) func() {
	ctx, cancel := context.WithCancel(context.Background())
	go func(ctx context.Context) {
		ts := time.Now()
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				es, _ := f.stackEvents(stack, &ts)
				for _, e := range es {
					printStackEvent(e)
				}
				if len(es) > 0 {
					ts = *es[0].Timestamp
				}
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}(ctx)
	return cancel
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
			return nil, log.Wrap(err)
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

func printStackEvent(e types.StackEvent) {
	var action string
	switch e.ResourceStatus {
	case types.ResourceStatusCreateComplete:
		action = "Created"
	case types.ResourceStatusDeleteComplete:
		action = "Destroyed"
	}
	if action == "" {
		return
	}
	fmt.Printf("\t%s %s %s", action, *e.ResourceType, *e.LogicalResourceId)
}
