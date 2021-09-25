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
	if v.Version == "" {
		return "latest"
	}
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

// TODO: make this point to some latest version
func (v *Version) functionsPath() string {
	if v.FunctionsPath == "" {
		return "functions"
	}
	return v.FunctionsPath
}
