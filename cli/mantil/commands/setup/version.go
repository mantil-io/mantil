package setup

import "fmt"

type Version struct {
	Commit        string
	Tag           string
	Dirty         bool
	Version       string
	FunctionsPath string
}

func (v *Version) String() string {
	return v.Version
}

func (v *Version) isPublished() bool {
	return v.Version == v.Tag
}

// published versions get replicated through the regions, dev ones are only located in central bucket
func (v *Version) setupBucket(region string) string {
	bucket := "mantil-downloads"
	if v.isPublished() {
		bucket = fmt.Sprintf("%s-%s", bucket, region)
	}
	return bucket
}
