package aws

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
)

func withRetry(callback func() error, isRetryable func(error) bool) error {
	retryInterval := time.Second
	retryAttempts := 60
	for retryAttempts > 0 {
		err := callback()
		if err == nil {
			break
		}
		if isRetryable(err) {
			time.Sleep(retryInterval)
			retryAttempts--
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func resourceFromARN(resourceARN string) (string, error) {
	if !arn.IsARN(resourceARN) {
		return "", fmt.Errorf("%s is not valid arn", resourceARN)
	}
	parts, err := arn.Parse(resourceARN)
	if err != nil {
		return "", fmt.Errorf("error parsing arn - %w", err)
	}
	return parts.Resource, nil
}
