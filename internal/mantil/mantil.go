package mantil

import (
	"fmt"

	"github.com/mantil-io/mantil/internal/aws"
)

const (
	s3SetupPrefix = "setup/"
)

func Bucket(aws *aws.AWS) (string, error) {
	accountID, err := aws.AccountID()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("mantil-%s-%s", aws.Region(), accountID), nil
}

func SetupBucketPrefix() string {
	return s3SetupPrefix
}

func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}
