package controller

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractApiStructNamePtr(t *testing.T) {
	stc, err := extractApiStructName("ping", "testdata/generate/ping_ptr")
	require.NoError(t, err)
	assert.Equal(t, "Ping", stc)
}

func TestExtractApiStructNameStruct(t *testing.T) {
	stc, err := extractApiStructName("ping", "testdata/generate/ping_struct")
	require.NoError(t, err)
	assert.Equal(t, "Ping", stc)
}

func TestExtractApiStructNameNoNew(t *testing.T) {
	_, err := extractApiStructName("ping", "testdata/generate/no_new")
	fmt.Println(err)
	require.Error(t, err)
}

func TestExtractApiStructNameNewWithParameters(t *testing.T) {
	_, err := extractApiStructName("ping", "testdata/generate/parameters")
	fmt.Println(err)
	require.Error(t, err)
}

func TestExtractApiStructNameTooManyValues(t *testing.T) {
	_, err := extractApiStructName("ping", "testdata/generate/return_values")
	fmt.Println(err)
	require.Error(t, err)
}

func TestExtractApiStructNameWrongValue(t *testing.T) {
	_, err := extractApiStructName("ping", "testdata/generate/return_value")
	fmt.Println(err)
	require.Error(t, err)
}

func TestExtractApiStructNameNewWrongType(t *testing.T) {
	_, err := extractApiStructName("ping", "testdata/generate/new_wrong_type")
	fmt.Println(err)
	require.Error(t, err)
}
