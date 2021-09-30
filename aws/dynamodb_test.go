package aws

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDynamodbTableNameFromARN(t *testing.T) {
	arn := "arn:aws:dynamodb:eu-central-1:158175150896:table/table-name"
	name, err := dynamodbTableNameFromARN(arn)
	require.NoError(t, err)
	assert.Equal(t, "table-name", name)

}
