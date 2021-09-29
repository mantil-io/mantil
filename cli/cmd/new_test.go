package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIsExternalRepo(t *testing.T) {
	cmd := &newCmd{}

	cmd.repo = "ping"
	assert.False(t, cmd.isExternalRepo())

	cmd.repo = "https://github.com/mantil-io/go-mantil-template"
	assert.True(t, cmd.isExternalRepo())

	cmd.repo = "git@github.com:mantil-io/go-mantil-template.git"
	assert.True(t, cmd.isExternalRepo())
}

func TestNewTemplate(t *testing.T) {
	cmd := &newCmd{}

	cmd.repo = "excuses"
	assert.Equal(t, "excuses", cmd.template())

	cmd.repo = "ping"
	assert.Equal(t, "ping", cmd.template())

	cmd.repo = ""
	assert.Equal(t, "ping", cmd.template())

	cmd.repo = "nonExistentTemplate"
	assert.Empty(t, cmd.template())
}

func TestNewRepoURL(t *testing.T) {
	cmd := &newCmd{}

	cmd.repo = "ping"
	url, err := cmd.repoURL()
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/mantil-io/go-mantil-template", url)

	cmd.repo = "nonExistentRepo"
	url, err = cmd.repoURL()
	require.Error(t, err)

	cmd.repo = "https://github.com/mantil-io/go-mantil-template"
	url, err = cmd.repoURL()
	require.NoError(t, err)
	assert.Equal(t, cmd.repo, url)
}
