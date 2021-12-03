package test

import (
	"testing"
)

func TestExcuses(t *testing.T) {
	r := newCliRunnerWithWorkspaceCopy(t)
	t.Parallel()

	projectName := "my-excuses"
	r.Assert("Your project is ready",
		"mantil", "new", projectName, "--from", "excuses")
	r.logf("created %s project in %s", projectName, r.SetWorkdir(projectName))

	r.Assert("Deploy successful!",
		"mantil", "stage", "new", "test", "--node", defaultNodeName)

	r.Assert("PASS",
		"mantil", "test")

	r.Assert(`"Excuse":`,
		"mantil", "invoke", "excuses/random")

	r.Assert(`"Count": 63`,
		"mantil", "invoke", "excuses/count")

	r.Assert(`204 No Content`,
		"mantil", "invoke", "excuses/clear")

	r.Assert(`no excuses`,
		"mantil", "invoke", "excuses/random")

	c := r.Assert(`count after: 109`,
		"mantil", "invoke", "excuses/load", "-d", `{"url":"https://gist.githubusercontent.com/orf/db8eb0aaddeea92dfcab/raw/5e9a8958fce65b1fe8f9bbaadeb87c207e5da848/gistfile1.txt"}`)
	r.StdoutContains(c, "count before: 0")
	r.StdoutContains(c, "λ")

	r.Assert(`"Count": 109`,
		"mantil", "invoke", "excuses/count")

	r.Assert("Stage test was successfully destroyed!",
		"mantil", "stage", "destroy", "test", "--yes")
}
