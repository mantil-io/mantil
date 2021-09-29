package aws

import (
	"time"
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
