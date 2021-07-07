package main

import (
	"path/filepath"
	"testing"

	"github.com/atoz-technology/mantil-cli/pkg/mantil"
	"github.com/stretchr/testify/require"
)

func setDevEnvPaths(t *testing.T) {
	var err error
	templatesFolder, err = filepath.Abs("../../templates")
	require.NoError(t, err)
	modulesFolder, err = filepath.Abs("../../..")
	require.NoError(t, err)
	// secretsFolder, err = filepath.Abs("../../../../secrets") //
	//require.NoError(t, err)

	// err = shell.PrepareHome(rootFolder+"/home", secretsFolder)
	// require.NoError(t, err)
}

func TestExample(t *testing.T) {
	setDevEnvPaths(t)

	p := &mantil.Project{}
	p.TestData()
	err := p.Apply()
	require.NoError(t, err)
}
