package terraform

import (
	"flag"
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/kit/shell"
	"github.com/mantil-io/mantil/node/dto"
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
		ResourceSuffix:  "abcdef",
		AuthEnv: map[string]string{
			"publicKey": "key",
		},
		ResourceTags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
	}
	tf, err := renderSetup(data)
	require.NoError(t, err)
	require.Equal(t, tf.createPath, "/tmp/mantil/setup/create")
	require.Equal(t, tf.destroyPath, "/tmp/mantil/setup/destroy")

	equalFiles(t, "testdata/setup.tf", "/tmp/mantil/setup/create/main.tf")
	equalFiles(t, "testdata/setup-destroy.tf", "/tmp/mantil/setup/destroy/main.tf")
}

func TestRenderProject(t *testing.T) {
	data := dto.StageTemplate{
		Project:             "my-project",
		Stage:               "my-stage",
		Bucket:              "bucket-name",
		BucketPrefix:        "bucket-prefix",
		NodeFunctionsBucket: "functions-bucket",
		NodeFunctionsPath:   "functions-path",
		Region:              "aws-region",
		ResourceSuffix:      "abcdef",
		Functions: []dto.Function{
			{
				Name:      "function1",
				S3Key:     "function1.zip",
				IsDefault: true,
			},
			{
				Name:  "function2",
				S3Key: "function2.zip",
			},
		},
		ResourceTags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
		WsEnv: map[string]string{
			"key": "value",
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

// func TestParseLog(t *testing.T) {
// 	content, err := ioutil.ReadFile("testdata/terraform_apply_output.txt")
// 	require.NoError(t, err)
// 	p := NewLogParser()

// 	scanner := bufio.NewScanner(strings.NewReader(string(content)))
// 	var parsedApply string
// 	for scanner.Scan() {
// 		line := scanner.Text()
// 		if l, _ := p.Parse(line); l != "" {
// 			if testing.Verbose() {
// 				fmt.Println(l)
// 			}
// 			parsedApply = parsedApply + l + "\n"
// 		}
// 	}
// 	if *update {
// 		err = os.WriteFile("testdata/terraform_apply_parsed.txt", []byte(parsedApply), 0644)
// 		require.NoError(t, err)
// 	}
// 	expectedApply, err := ioutil.ReadFile("testdata/terraform_apply_parsed.txt")
// 	require.NoError(t, err)
// 	require.Equal(t, parsedApply, string(expectedApply))

// 	require.Len(t, p.Outputs, 4)
// 	require.Equal(t, p.Outputs["functions_bucket"], "mantil-eu-central-1-4byktiy")
// 	require.Equal(t, p.Outputs["public_site_bucket"], "mantil-public-project-dev-4byktiy")
// 	require.Equal(t, p.Outputs["url"], "https://y3z3mojmrk.execute-api.eu-central-1.amazonaws.com")
// 	require.Equal(t, p.Outputs["ws_url"], "wss://rjj99anpxg.execute-api.eu-central-1.amazonaws.com")

// 	content, err = ioutil.ReadFile("testdata/terraform_destroy.txt")
// 	require.NoError(t, err)
// 	scanner = bufio.NewScanner(strings.NewReader(string(content)))
// 	var parsed string
// 	for scanner.Scan() {
// 		if l, _ := p.Parse(scanner.Text()); l != "" {
// 			if testing.Verbose() {
// 				fmt.Println(l)
// 			}
// 			parsed = parsed + l + "\n"
// 		}
// 	}
// 	if *update {
// 		err = os.WriteFile("testdata/terraform_destroy_parsed.txt", []byte(parsed), 0644)
// 		require.NoError(t, err)
// 	}
// 	expected, err := ioutil.ReadFile("testdata/terraform_destroy_parsed.txt")
// 	require.NoError(t, err)
// 	require.Equal(t, parsed, string(expected))

// }
