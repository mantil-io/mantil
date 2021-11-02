// Package build collects build time information
// Decides where is the deployment location; s3 bucket and key
// for lambda function packages.
// That information is used in script/deploy.sh to put build artifacts in the right place.
package build

import (
	_ "embed"
	"fmt"
)

const (
	// bucket where we upload deployments
	releasesBucket = "mantil-releases"
)

// embeded credentials file for access to ngs for publishing mantil events
//go:embed event-publisher.creds
var EventPublisherCreds string

// global variables set in the build time by ld flags
var (
	tag   string
	dev   string
	ontag string
)

// current deployment version description
func Version() string {
	return newDeploymentInfo(tag, dev, ontag).String()
}

// deployment information
// where to put/get deployment artifacts (lambda function packages)
func Deployment() DeploymentInfo {
	return newDeploymentInfo(tag, dev, ontag)
}

// Collects build time information.
// Decides where is the deployment location; s3 bucket and key
// for lambda function pacakges.
type DeploymentInfo struct {
	tag     string
	dev     string
	release bool
}

func newDeploymentInfo(tag, dev, onTag string) DeploymentInfo {
	return DeploymentInfo{
		tag:     tag,
		release: (onTag != "" && onTag != "0"),
		dev:     dev,
	}
}

// is this release or development deployment
func (v DeploymentInfo) Release() bool {
	return v.release
}

// current deployment description
func (v DeploymentInfo) String() string {
	return v.tag
}

// PutPath s3 path where we deploy releases
// it is always in releaseBucket
// we are replicating this bucket to other regional buckets
func (v DeploymentInfo) PutPath() string {
	return fmt.Sprintf("s3://%s/%s/", releasesBucket, v.bucketKey())
}

func (v DeploymentInfo) LatestBucket() string {
	if v.release {
		return fmt.Sprintf("s3://%s/latest/", releasesBucket)
	}
	return ""
}

// GetPath returns bucket and key in the bucket for reading deployed functions
func (v DeploymentInfo) GetPath(region string) (string, string) {
	return v.getBucket(region), v.bucketKey()
}

// bucket for reading functions
// it has to be in the same region as lambda functions which are created from uploaded resources
// so it is different from deploy bucket
func (v DeploymentInfo) getBucket(region string) string {
	if v.release && region != "" {
		return fmt.Sprintf("%s-%s", releasesBucket, region)
	}
	return releasesBucket
}

// key inside deploy or replicated bucket
// it is same in both cases
func (v DeploymentInfo) bucketKey() string {
	if v.release {
		return fmt.Sprintf("%s", v.tag)
	}
	return fmt.Sprintf("dev/%s/%s", v.dev, v.tag)
}
