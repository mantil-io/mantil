package build

import (
	"fmt"
	"os"
)

const (
	contextKey          = "version"
	functionsPathEnv    = "MANTIL_TESTS_FUNCTIONS_PATH"
	functionsBucketName = "mantil-releases"
)

var (
	tag   string
	dev   string
	ontag string

	version VersionInfo
)

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

func (v *VersionInfo) uploadPath() string {
	if v.release {
		return fmt.Sprintf("%s", v.tag)
	}
	return fmt.Sprintf("dev/%s/%s", v.dev, v.tag)
}

func (v VersionInfo) String() string {
	return v.tag
}

func (v *VersionInfo) UploadBucket() string {
	return fmt.Sprintf("s3://%s/%s/", functionsBucketName, v.uploadPath())
}

func (v *VersionInfo) LatestBucket() string {
	if v.release {
		return fmt.Sprintf("s3://%s/latest/", functionsBucketName)
	}
	return ""
}

func (v *VersionInfo) Release() bool {
	return v.release
}

func Log() string {
	return fmt.Sprintf("tag: %s, dev: %s, ontag: %s", tag, dev, ontag)
}

func (v *VersionInfo) FunctionsBucket(region string) string {
	if v.release {
		return fmt.Sprintf("%s-%s", functionsBucketName, region)
	}
	return functionsBucketName
}

func (v *VersionInfo) FunctionsPath() string {
	if v.tag == "" {
		if val, ok := os.LookupEnv(functionsPathEnv); ok {
			return val
		}
		return "functions/latest"
	}
	return v.uploadPath()
}

func Version() VersionInfo {
	return newVersion(tag, dev, ontag)
}
