package setup

import (
	"io/ioutil"
	"testing"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/stretchr/testify/require"
)

func TestRenderTemplate(t *testing.T) {
	td := stackTemplateData{
		Name:   "mantil-setup",
		Bucket: "bucket",
		S3Key:  "bucket-key",
		Region: "region",
	}
	actual, err := renderStackTemplate(td)
	require.NoError(t, err)
	expected, err := ioutil.ReadFile("testdata/template.yml")
	require.NoError(t, err)
	equalStrings(t, string(expected), actual)
}

func equalStrings(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Logf("diff of strings")
		dmp := diffmatchpatch.New()
		diffs := dmp.DiffMain(expected, actual, false)
		t.Logf("diff: \n%s", dmp.DiffPrettyText(diffs))
		t.Fatalf("failed")
	}
}
