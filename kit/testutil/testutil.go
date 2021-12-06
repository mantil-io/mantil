package testutil

import (
	"io/ioutil"
	"testing"

	"github.com/mantil-io/mantil/kit/shell"
)

func EqualFiles(t *testing.T, expected, actual string, update bool) {
	actualContent, err := ioutil.ReadFile(actual)
	if err != nil {
		t.Fatalf("failed reading actual file: %s", err)
	}

	if update {
		t.Logf("update expected file %s", expected)
		if err := ioutil.WriteFile(expected, actualContent, 0644); err != nil {
			t.Fatalf("failed to update expectexd file: %s", err)
		}
		return
	}

	expectedContent, err := ioutil.ReadFile(expected)
	if err != nil {
		t.Fatalf("failed reading expected file: %s", err)
	}

	if string(actualContent) != string(expectedContent) {
		args := []string{"diff", expected, actual}
		out, err := shell.Output(shell.ExecOptions{Args: args})
		if err != nil {
			t.Logf("diff of files")
			t.Logf("expected %s, actual %s", expected, actual)
			t.Logf("%s", out)
			t.Fatalf("failed")
		}

	}
}
