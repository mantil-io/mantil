package setup

import (
	"fmt"
	"os"
)

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
	if v.Version == "" && v.Tag == "" {
		return false
	}
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

const functionsPathEnv = "MANTIL_TESTS_FUNCTIONS_PATH"

// TODO: make this point to some latest version
func (v *Version) functionsPath() string {
	if v.FunctionsPath == "" {
		if val, ok := os.LookupEnv(functionsPathEnv); ok {
			return val
		}
		return "functions/latest"
	}
	return v.FunctionsPath
}
