package setup

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/terraform"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update expected files")

func TestTerraformRender(t *testing.T) {
	data := terraform.SetupTemplateData{
		Bucket:          "bucket-name",
		BucketPrefix:    "bucket-prefix",
		FunctionsBucket: "functions-bucket",
		FunctionsPath:   "functions-path",
		Region:          "aws-region",
		PublicKey:       "public-key",
		ResourceSuffix:  "abcdef",
		ResourceTags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	tf, err := terraform.Setup(data)
	require.NoError(t, err)
	equalFiles(t, "./testdata/create.tf", tf.CreateTf())
	equalFiles(t, "./testdata/destroy.tf", tf.DestroyTf())
}

// TODO: same function in terraform package
func equalFiles(t *testing.T, expected, actual string) {
	actualContent, err := ioutil.ReadFile(actual)
	if err != nil {
		t.Fatalf("failed reading actual file: %s", err)
	}

	if *update {
		t.Logf("update expected file %s", expected)
		if err := ioutil.WriteFile(expected, actualContent, 0644); err != nil {
			t.Fatalf("failed to update expectexd file: %s", err)
		}
		return
	}

	expectedContent, err := ioutil.ReadFile(expected)
	if err != nil {
		t.Fatalf("failed reading expected file: %s", err)
	}

	if string(actualContent) != string(expectedContent) {
		args := []string{"diff", expected, actual}
		out, err := shell.Output(shell.ExecOptions{Args: args})
		if err != nil {
			t.Logf("diff of files")
			t.Logf("expected %s, actual %s", expected, actual)
			t.Logf("%s", out)
			t.Fatalf("failed")
		}

	}
}
