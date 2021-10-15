package workspace

import (
	"fmt"

	"github.com/mantil-io/mantil/aws"
)

func Bucket(aws *aws.AWS, resourceSuffix string) string {
	return bucket(aws.Region(), resourceSuffix)
}

func bucket(awsRegion, resourceSuffix string) string {
	return fmt.Sprintf("mantil-%s-%s", awsRegion, resourceSuffix)
}

func RuntimeResource(v ...string) string {
	r := "mantil"
	for _, n := range v {
		r = fmt.Sprintf("%s-%s", r, n)
	}
	return r
}
