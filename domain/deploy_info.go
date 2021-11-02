package domain

import (
	"fmt"
)

const (
	// bucket where we upload deployments
	releasesBucket = "mantil-releases"
)

// global used in Version and Deployment
var deployInfo DeployInfo

// global variables set in the build time by ld flags
var (
	tag   string
	dev   string
	ontag string
)

// current deployment version description
func Version() string {
	return Deployment().String()
}

// deployment information
// where to put/get deployment artifacts (lambda function packages)
func Deployment() DeployInfo {
	return newDeployInfo(tag, dev, ontag)
}

// Collects build time information. Decides where is the deployment location; s3
// bucket and key for lambda function pacakges. That information is used in
// script/deploy.sh to put build artifacts in the right place.
type DeployInfo struct {
	tag     string
	dev     string
	release bool
}

func newDeployInfo(tag, dev, onTag string) DeployInfo {
	return DeployInfo{
		tag:     tag,
		release: (onTag != "" && onTag != "0"),
		dev:     dev,
	}
}

// is this release or development deployment
func (d DeployInfo) Release() bool {
	return d.release
}

// current deployment description
func (d DeployInfo) String() string {
	return d.tag
}

// PutPath s3 path where we deploy releases
// it is always in releaseBucket
// we are replicating this bucket to other regional buckets
func (d DeployInfo) PutPath() string {
	return fmt.Sprintf("s3://%s/%s/", releasesBucket, d.bucketKey())
}

func (d DeployInfo) LatestBucket() string {
	if d.release {
		return fmt.Sprintf("s3://%s/latest/", releasesBucket)
	}
	return ""
}

// GetPath returns bucket and key in the bucket for reading deployed functions
func (d DeployInfo) GetPath(region string) (string, string) {
	return d.getBucket(region), d.bucketKey()
}

// bucket for reading functions
// it has to be in the same region as lambda functions which are created from uploaded resources
// so it is different from deploy bucket
func (d DeployInfo) getBucket(region string) string {
	if d.release && region != "" {
		return fmt.Sprintf("%s-%s", releasesBucket, region)
	}
	return releasesBucket
}

// key inside deploy or replicated bucket
// it is same in both cases
func (d DeployInfo) bucketKey() string {
	if d.release {
		return fmt.Sprintf("%s", d.tag)
	}
	return fmt.Sprintf("dev/%s/%s", d.dev, d.tag)
}
