package setup

import (
	"context"
	"fmt"
	"os"
)

const (
	contextKey          = "version"
	functionsPathEnv    = "MANTIL_TESTS_FUNCTIONS_PATH"
	functionsBucketName = "mantil-downloads"
)

type VersionInfo struct {
	tag     string
	dev     string
	release bool
}

func NewVersion(tag, dev, onTag string) *VersionInfo {
	return &VersionInfo{
		tag:     tag,
		release: (onTag != "" && onTag != "0"),
		dev:     dev,
	}
}

func (v *VersionInfo) uploadPath() string {
	if v.release {
		return fmt.Sprintf("functions/%s", v.tag)
	}
	return fmt.Sprintf("dev/%s/%s", v.dev, v.tag)
}

func (v *VersionInfo) String() string {
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

func (v *VersionInfo) functionsBucket(region string) string {
	if v.release {
		return fmt.Sprintf("%s-%s", functionsBucketName, region)
	}
	return functionsBucketName
}

func (v *VersionInfo) functionsPath() string {
	if v.tag == "" {
		if val, ok := os.LookupEnv(functionsPathEnv); ok {
			return val
		}
		return "functions/latest"
	}
	return v.uploadPath()
}

func GetVersion(ctx context.Context) (*VersionInfo, bool) {
	lc, ok := ctx.Value(contextKey).(*VersionInfo)
	return lc, ok
}

func SetVersion(ctx context.Context, v *VersionInfo) context.Context {
	return context.WithValue(ctx, contextKey, v)
}
