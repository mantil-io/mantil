package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogGroupNameFromARN(t *testing.T) {
	arn := "arn:aws:logs:eu-central-1:158175150896:log-group:/aws/lambda/name:*"
	name, err := logGroupNameFromARN(arn)
	require.NoError(t, err)
	assert.Equal(t, "/aws/lambda/name", name)
}
