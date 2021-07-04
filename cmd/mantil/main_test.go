package main

import (
	"path/filepath"
	"testing"

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

	spa := &Spa{}
	spa.testData()
	err := spa.Apply()
	require.NoError(t, err)
}
