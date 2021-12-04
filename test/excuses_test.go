package test

import (
	"testing"
)

func TestExcuses(t *testing.T) {
	c := newClitestWithWorkspaceCopy(t)
	t.Parallel()

	projectName := "my-excuses"
	c.Run("mantil", "new", projectName, "--from", "excuses").
		Contains("Your project is ready")
	t.Logf("created %s project in %s", projectName, c.Cd(projectName))

	c.Run("mantil", "stage", "new", "test", "--node", defaultNodeName).
		Contains("Deploy successful!")

	c.Run("mantil", "test").Contains("PASS")

	c.Run("mantil", "invoke", "excuses/random").
		Contains(`"Excuse":`)

	c.Run("mantil", "invoke", "excuses/count").
		Contains(`"Count": 63`)

	c.Run("mantil", "invoke", "excuses/clear").
		Contains(`204 No Content`)

	c.Run("mantil", "invoke", "excuses/random").
		Contains(`no excuses`)

	c.Run("mantil", "invoke", "excuses/load", "-d", `{"url":"https://gist.githubusercontent.com/orf/db8eb0aaddeea92dfcab/raw/5e9a8958fce65b1fe8f9bbaadeb87c207e5da848/gistfile1.txt"}`).
		Contains(`count after: 109`).
		Contains("count before: 0").
		Contains("λ")

	c.Run("mantil", "invoke", "excuses/count").
		Contains(`"Count": 109`)

	c.Run("mantil", "stage", "destroy", "test", "--yes").
		Contains("Stage test was successfully destroyed!")
}
