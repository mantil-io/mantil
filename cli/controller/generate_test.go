package controller

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsNewValidPtr(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/ping_ptr")
	require.NoError(t, err)
}

func TestIsNewValidStruct(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/ping_struct")
	require.NoError(t, err)
}

func TestIsNewValidNoNew(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/no_new")
	require.Error(t, err)
}

func TestIsNewValidWithParameters(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/parameters")
	require.Error(t, err)
}

func TestIsNewValidTooManyReturnValues(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/return_values")
	require.Error(t, err)
}

func TestIsNewValidWrongReturnValue(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/return_value")
	require.Error(t, err)
}

func TestIsNewValidNewWrongType(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/new_wrong_type")
	require.Error(t, err)
}

func TestIsNewValidMethod(t *testing.T) {
	err := isNewValid("ping", "testdata/generate/new_method")
	require.Error(t, err)
}
