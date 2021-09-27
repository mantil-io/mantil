package setup

import (
	"flag"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/terraform"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update expected files")

func TestInit(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: AWS client not initialized")
	}
	init := func() *Setup {
		req := dto.SetupRequest{}
		s := New()
		err := s.init(&req, cli)
		require.NoError(t, err)
		return s
	}

	t.Run("init", func(t *testing.T) {
		s := init()
		require.NotNil(t, s.tf)
		require.NotNil(t, s.awsClient)
		id, err := s.awsClient.AccountID()
		require.NoError(t, err)
		require.True(t, strings.HasSuffix(s.bucketName, id))

		t.Logf("accountID: %s\n", id)
		t.Logf("bucket: %s\n", s.bucketName)
	})

}

func TestTerraformRender(t *testing.T) {
	tf, err := terraform.New("mantil-setup")
	require.NoError(t, err)
	data := terraform.SetupTemplateData{
		Bucket:          "bucket-name",
		BucketPrefix:    "bucket-prefix",
		FunctionsBucket: "functions-bucket",
		FunctionsPath:   "functions-path",
		Region:          "aws-region",
		PublicKey:       "public-key",
	}
	err = tf.RenderSetupTemplate(data)
	require.NoError(t, err)
	requireEqual(t, "./testdata/mantil-setup-main.tf", "/tmp/mantil-setup/main.tf")
}

func requireEqual(t *testing.T, expected, actual string) {
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
		out, err := shell.Output(args, "")
		if err != nil {
			t.Logf("diff of files")
			t.Logf("expected %s, actual %s", expected, actual)
			t.Logf("%s", out)
			t.Fatalf("failed")
		}

	}
}
