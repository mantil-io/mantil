package workspace

import (
	"fmt"

	"github.com/mantil-io/mantil/aws"
)

func Bucket(aws *aws.AWS) string {
	return fmt.Sprintf("mantil-%s-%s", aws.Region(), aws.AccountID())
}

func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}
