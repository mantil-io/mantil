package build

import (
	_ "embed"
	"fmt"
)

const (
	// bucket where we upload deployments
	releasesBucket = "mantil-releases"
)

//go:embed event-publisher.creds
var EventPublisherCreds string

// global variables set in the build time by ld flags
var (
	tag   string
	dev   string
	ontag string
)

func Version() VersionInfo {
	return newVersion(tag, dev, ontag)
}

// Collects build time information.
// Descides where is the deployment location; s3 bucket and key
// for node functions.
type VersionInfo struct {
	tag     string
	dev     string
	release bool
}

func newVersion(tag, dev, onTag string) VersionInfo {
	return VersionInfo{
		tag:     tag,
		release: (onTag != "" && onTag != "0"),
		dev:     dev,
	}
}

// is this release or development deplolyment
func (v *VersionInfo) Release() bool {
	return v.release
}

// current version description
func (v VersionInfo) String() string {
	return v.tag
}

// DeployPath s3 path where we deploy releases
// it is always in releaseBucket
// we are replicating this bucket to other regional buckets
func (v *VersionInfo) DeployPath() string {
	return fmt.Sprintf("s3://%s/%s/", releasesBucket, v.bucketKey())
}

func (v *VersionInfo) LatestBucket() string {
	if v.release {
		return fmt.Sprintf("s3://%s/latest/", releasesBucket)
	}
	return ""
}

// GetPath returns bucket and key in the bucket for reading deployed functions
func (v VersionInfo) GetPath(region string) (string, string) {
	return v.getBucket(region), v.bucketKey()
}

// bucket for reading functions
// it has to be in the same region as lambda functions which are created from uploaded resources
// so it is different from deploy bucket
func (v *VersionInfo) getBucket(region string) string {
	if v.release && region != "" {
		return fmt.Sprintf("%s-%s", releasesBucket, region)
	}
	return releasesBucket
}

// key inside deploy or replicated bucket
// it is same in both cases
func (v *VersionInfo) bucketKey() string {
	if v.release {
		return fmt.Sprintf("%s", v.tag)
	}
	return fmt.Sprintf("dev/%s/%s", v.dev, v.tag)
}
