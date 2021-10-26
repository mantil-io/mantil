package controller

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIsExternalRepo(t *testing.T) {
	assert.False(t, isExternalRepo("repo"))
	assert.True(t, isExternalRepo("https://github.com/mantil-io/go-mantil-template"))
	assert.True(t, isExternalRepo("git@github.com:mantil-io/go-mantil-template.git"))
}

func TestNewProjectTemplate(t *testing.T) {
	assert.Equal(t, "excuses", projectTemplate("excuses"))
	assert.Equal(t, "ping", projectTemplate("ping"))
	assert.Equal(t, "ping", projectTemplate(""))
	assert.Empty(t, projectTemplate("nonExistentTemplate"))
}

func TestNewRepoURL(t *testing.T) {
	url, err := repoURL("name", "ping")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/mantil-io/go-mantil-template", url)

	url, err = repoURL("name", "nonExistentRepo")
	require.Error(t, err)

	url, err = repoURL("name", "https://github.com/mantil-io/go-mantil-template")
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/mantil-io/go-mantil-template", url)
}
