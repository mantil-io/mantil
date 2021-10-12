package setup

import (
	"flag"
	"io/ioutil"
	"log"
	"testing"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/aws"
	"github.com/mantil-io/mantil/shell"
	"github.com/mantil-io/mantil/terraform"
	"github.com/mantil-io/mantil/workspace"
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
	}
	tf, err := terraform.Setup(data)
	require.NoError(t, err)
	equalFiles(t, "./testdata/create.tf", tf.CreateTf())
	equalFiles(t, "./testdata/destroy.tf", tf.DestroyTf())
}

func TestIntegration(t *testing.T) {
	cli := aws.NewForTests(t)
	if cli == nil {
		t.Skip("skip: AWS client not initialized")
	}
	s := func() *Setup {
		req := dto.SetupRequest{
			Bucket:          workspace.Bucket(cli),
			FunctionsBucket: "mantil-downloads",
			FunctionsPath:   "functions/latest",
			PublicKey:       "my-test-public-key",
		}
		s := New()
		err := s.init(&req, cli)
		require.NoError(t, err)
		return s
	}()
	t.Run("init", func(t *testing.T) {
		require.NotNil(t, s.awsClient)
	})

	t.Run("create bucket", func(t *testing.T) {
		require.NoError(t, s.createBucket())
	})
	t.Run("create", func(t *testing.T) {
		out, err := s.terraformCreate()
		require.NoError(t, err)
		require.NotEmpty(t, out.APIGatewayRestURL)
		require.NotEmpty(t, out.APIGatewayWsURL)
		log.Printf("output: %#v", out)
	})
	t.Run("destroy", func(t *testing.T) {
		err := s.terraformDestroy()
		require.NoError(t, err)
	})
	t.Run("delete bucket", func(t *testing.T) {
		require.NoError(t, s.deleteBucket())
	})
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
