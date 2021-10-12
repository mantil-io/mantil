package destroy

import (
	"testing"

	"github.com/mantil-io/mantil/api/dto"
	"github.com/mantil-io/mantil/workspace"
	"github.com/stretchr/testify/assert"
)

func TestTerraformProjectTemplateData(t *testing.T) {
	d := &Destroy{
		req: &dto.DestroyRequest{
			Bucket:      "bucket",
			ProjectName: "test-project",
			StageName:   "test-stage",
		},
		stage: &workspace.Stage{
			Name: "test-stage",
		},
		region: "region,",
	}
	data := d.terraformProjectTemplateData()
	assert.NotEmpty(t, data.Name)
	assert.NotEmpty(t, data.Bucket)
	assert.NotEmpty(t, data.BucketPrefix)
	assert.NotEmpty(t, data.Region)
	assert.NotEmpty(t, data.Stage)
}
