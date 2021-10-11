package workspace

import (
	"fmt"

	"github.com/mantil-io/mantil/aws"
)

func Bucket(aws *aws.AWS) string {
	return bucket(aws.Region(), aws.AccountID())
}

func bucket(awsRegion, awsAccountID string) string {
	return fmt.Sprintf("mantil-%s-%s", awsRegion, awsAccountID)
}

func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}
