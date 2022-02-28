package setup

import (
	"flag"
	"testing"

	"github.com/mantil-io/mantil/kit/testutil"
	"github.com/mantil-io/mantil/node/terraform"
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
		ResourceSuffix:  "abcdef",
		AuthEnv: map[string]string{
			"publicKey": "public_key",
		},
		ResourceTags: map[string]string{
			"tag1": "value1",
			"tag2": "value2",
		},
		NamingTemplate: "mantil-%s",
		PublicKey:      "public_key",
		PrivateKey:     "private_key",
		GithubOrg:      "github_org",
	}
	tf, err := terraform.Setup(data)
	require.NoError(t, err)
	testutil.EqualFiles(t, "./testdata/create.tf", tf.CreateTf(), *update)
	testutil.EqualFiles(t, "./testdata/destroy.tf", tf.DestroyTf(), *update)
}
