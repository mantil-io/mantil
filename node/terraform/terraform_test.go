package terraform

import (
	"bufio"
	"flag"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/mantil-io/mantil/kit/testutil"
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
		NamingTemplate: "prefix-%s-suffix",
	}
	tf, err := renderSetup(data)
	require.NoError(t, err)
	require.Equal(t, tf.createPath, "/tmp/mantil/setup/create")
	require.Equal(t, tf.destroyPath, "/tmp/mantil/setup/destroy")

	testutil.EqualFiles(t, "testdata/setup.tf", "/tmp/mantil/setup/create/main.tf", *update)
	testutil.EqualFiles(t, "testdata/setup-destroy.tf", "/tmp/mantil/setup/destroy/main.tf", *update)
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
				Name:  "function1",
				S3Key: "function1.zip",
				Cron:  "* * * * ? *",
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
		HasPublic:        true,
		NamingTemplate:   "prefix-%s-suffix",
		PublicBucketName: "public-bucket",
		CustomDomain: dto.CustomDomain{
			DomainName:       "example.com",
			CertDomain:       "example.com",
			HostedZoneDomain: "example.com",
			HttpSubdomain:    "",
			WsSubdomain:      "ws",
		},
	}
	tf, err := renderProject(data)
	require.NoError(t, err)
	require.Equal(t, tf.createPath, "/tmp/mantil/my-project-my-stage/create")
	require.Equal(t, tf.destroyPath, "/tmp/mantil/my-project-my-stage/destroy")

	testutil.EqualFiles(t, "testdata/project.tf", "/tmp/mantil/my-project-my-stage/create/main.tf", *update)
	testutil.EqualFiles(t, "testdata/project-destroy.tf", "/tmp/mantil/my-project-my-stage/destroy/main.tf", *update)
}

func TestParseLog(t *testing.T) {
	p := NewLogParser()
	testStateChanges(t, p, "testdata/terraform_apply_output.txt", map[int]ParserState{
		1:   StateInitializing,
		765: StateCreating,
		852: StateDone,
	})
	require.Nil(t, p.Error())
	require.Len(t, p.Outputs, 4)
	require.Equal(t, "mantil-aef7a9da", p.Outputs["functions_bucket"])
	require.Equal(t, "", p.Outputs["public_site_bucket"])
	require.Equal(t, "https://9mosxdgpy2.execute-api.eu-central-1.amazonaws.com", p.Outputs["url"])
	require.Equal(t, "wss://976orve3jg.execute-api.eu-central-1.amazonaws.com", p.Outputs["ws_url"])

	p = NewLogParser()
	testStateChanges(t, p, "testdata/terraform_destroy_output.txt", map[int]ParserState{
		1:    StateInitializing,
		1384: StateDestroying,
		1455: StateDone,
	})
	require.Nil(t, p.Error())

	p = NewLogParser()
	testStateChanges(t, p, "testdata/terraform_error_output.txt", map[int]ParserState{
		1:   StateInitial,
		690: StateUpdating,
		701: StateDone,
	})
	require.NotNil(t, p.Error())
}

func testStateChanges(t *testing.T, p *Parser, dataPath string, stateChanges map[int]ParserState) {
	content, err := ioutil.ReadFile(dataPath)
	require.NoError(t, err)
	require.Equal(t, StateInitial, p.State())

	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	currentState := StateInitial
	lineCnt := 0
	for scanner.Scan() {
		lineCnt++
		line := scanner.Text()
		isTf := p.Parse(line)
		require.True(t, isTf, line)
		s, ok := stateChanges[lineCnt]
		if ok {
			currentState = s
		}
		require.Equal(t, p.State(), currentState)
	}
}
