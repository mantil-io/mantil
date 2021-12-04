package clitest

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSuccess(t *testing.T) {
	e := New(t)
	e.Run("true").Success()
}

func TestFail(t *testing.T) {
	e := New(t)
	e.Run("false").Fail()
}

func TestContains(t *testing.T) {
	e := New(t)
	e.Run("echo", "jozo bozo").Stdout().Contains("jozo").Contains("bozo")
	out, err := ioutil.ReadFile(e.stdoutFilename)
	require.NoError(t, err)
	require.Equal(t, "jozo bozo\n", string(out))
}

func TestLongRunning(t *testing.T) {
	t.Skip()
	e := New(t)
	e.Run("./long_running.sh").Success().Stdout().Contains("1").Contains("2")
}

func TestEnv(t *testing.T) {
	t.Setenv("pero", "zdero")
	e := New(t)
	require.NotEmpty(t, e.vars)
	require.Contains(t, e.vars, "pero=zdero")

	e.Env("pero", "bozo")
	require.NotContains(t, e.vars, "pero=zdero")
	require.Contains(t, e.vars, "pero=bozo")
}

func TestVarsWithout(t *testing.T) {
	t.Setenv("pero", "zdero")
	e := New(t)
	vars, val := e.varsWithout("pero")
	require.NotContains(t, vars, "pero=zdero")
	require.Equal(t, "zdero", val)
}

func TestPath(t *testing.T) {
	e := New(t)
	_, pathBefore := e.varsWithout("PATH")
	e.Path("/tmp")

	_, pathAfter := e.varsWithout("PATH")
	require.True(t, strings.HasPrefix(pathAfter, "/tmp:"))
	require.True(t, strings.HasSuffix(pathAfter, pathBefore))
}
