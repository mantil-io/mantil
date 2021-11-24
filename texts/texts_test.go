package texts

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActivationMailBody(t *testing.T) {
	content, err := ActivationMailBody("ianic", "1234567890")
	require.NoError(t, err)
	//t.Logf("content:\n%s", content)
	require.True(t, strings.Contains(content, "ianic"))
	require.True(t, strings.Contains(content, "1234567890"))
}

func TestHTMLActivationMail(t *testing.T) {
	content, err := ActivationHTMLMailBody("Ksenija", "3UTZElNhReugz-3JVfU4nQ")
	require.NoError(t, err)
	file, err := ioutil.TempFile("", "*.html")
	fmt.Fprint(file, content)
	t.Logf("created: %s\n", file.Name())
	exec.Command("open", file.Name()).Start()
}

func TestHTMLWelcomeMail(t *testing.T) {
	content, err := WelcomeMailHTMLBody("Ksenija")
	require.NoError(t, err)
	file, err := ioutil.TempFile("", "*.html")
	fmt.Fprint(file, content)
	t.Logf("created: %s\n", file.Name())
	exec.Command("open", file.Name()).Start()
}
