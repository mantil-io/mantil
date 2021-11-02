package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDevelopmentVersion(t *testing.T) {
	v := newDeployInfo("v0.1.13-27-gc27a51c", "ianic", "")

	require.False(t, v.Release())
	require.Equal(t, v.bucketKey(), "dev/ianic/v0.1.13-27-gc27a51c")
	require.Equal(t, v.PutPath(), "s3://mantil-releases/dev/ianic/v0.1.13-27-gc27a51c/")
	require.Equal(t, v.LatestBucket(), "")

	require.Equal(t, v.getBucket(""), "mantil-releases")
	require.Equal(t, v.getBucket("us-east-1"), "mantil-releases")
	bucket, key := v.GetPath("us-east-1")
	require.Equal(t, bucket, "mantil-releases")
	require.Equal(t, key, "dev/ianic/v0.1.13-27-gc27a51c")
}

func TestReleaseVersion(t *testing.T) {
	v := newDeployInfo("v0.1.13", "ianic", "1")

	require.True(t, v.Release())
	require.Equal(t, v.bucketKey(), "v0.1.13")
	require.Equal(t, v.PutPath(), "s3://mantil-releases/v0.1.13/")
	require.Equal(t, v.LatestBucket(), "s3://mantil-releases/latest/")

	require.Equal(t, v.getBucket(""), "mantil-releases")
	require.Equal(t, v.getBucket("us-east-1"), "mantil-releases-us-east-1")

	bucket, key := v.GetPath("us-east-1")
	require.Equal(t, bucket, "mantil-releases-us-east-1")
	require.Equal(t, key, "v0.1.13")
}
