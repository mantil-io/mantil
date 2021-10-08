package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResourceFromARN(t *testing.T) {
	invalidARN := "invalidARN"
	_, err := resourceFromARN(invalidARN)
	require.Error(t, err)

	lgARN := "arn:aws:logs:eu-central-1:158175150896:log-group:/aws/lambda/name:*"
	lgResource, err := resourceFromARN(lgARN)
	require.NoError(t, err)
	assert.Equal(t, "log-group:/aws/lambda/name:*", lgResource)

	dbARN := "arn:aws:dynamodb:eu-central-1:158175150896:table/table-name"
	dbResource, err := resourceFromARN(dbARN)
	require.NoError(t, err)
	assert.Equal(t, "table/table-name", dbResource)
}
