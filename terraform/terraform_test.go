package terraform

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/config"
	"github.com/mantil-io/mantil/shell"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update expected files")

func TestEmbed(t *testing.T) {
	modules, err := fs.ReadDir("modules")
	require.NoError(t, err)
	require.NotEmpty(t, modules)

	templates, err := fs.ReadDir("templates")
	require.NoError(t, err)
	require.NotEmpty(t, templates)
}

func TestRenderSetup(t *testing.T) {
	data := SetupTemplateData{
		Bucket:          "bucket-name",
		BucketPrefix:    "bucket-prefix",
		FunctionsBucket: "functions-bucket",
		FunctionsPath:   "functions-path",
		Region:          "aws-region",
		PublicKey:       "public-key",
	}
	tf, err := renderSetup(data)
	require.NoError(t, err)
	require.Equal(t, tf.createPath, "/tmp/mantil/setup/create")
	require.Equal(t, tf.destroyPath, "/tmp/mantil/setup/destroy")

	equalFiles(t, "testdata/setup.tf", "/tmp/mantil/setup/create/main.tf")
	equalFiles(t, "testdata/setup-destroy.tf", "/tmp/mantil/setup/destroy/main.tf")
}

func TestRenderProject(t *testing.T) {
	data := ProjectTemplateData{
		Name:                   "my-project",
		Stage:                  "my-stage",
		Bucket:                 "bucket-name",
		BucketPrefix:           "bucket-prefix",
		RuntimeFunctionsBucket: "functions-bucket",
		RuntimeFunctionsPath:   "functions-path",
		Region:                 "aws-region",
		Functions: []*config.Function{
			{
				Name:  "function1",
				S3Key: "function1.zip",
			},
			{
				Name:  "function2",
				S3Key: "function2.zip",
			},
		},
		GlobalEnv: map[string]string{
			"env1": "value1",
			"env2": "value2",
		},
	}
	tf, err := renderProject(data)
	require.NoError(t, err)
	require.Equal(t, tf.createPath, "/tmp/mantil/my-project-my-stage/create")
	require.Equal(t, tf.destroyPath, "/tmp/mantil/my-project-my-stage/destroy")

	equalFiles(t, "testdata/project.tf", "/tmp/mantil/my-project-my-stage/create/main.tf")
	equalFiles(t, "testdata/project-destroy.tf", "/tmp/mantil/my-project-my-stage/destroy/main.tf")
}

// TODO: same function in api/setup package
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
