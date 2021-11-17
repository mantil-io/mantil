package texts

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestActivationMailBody(t *testing.T) {
	content, err := ActivationMailBody("ianic", "1234567890")
	require.NoError(t, err)
	t.Logf("content:\n%s", content)

	expected := `Hi ianic,

Your activation token is: 1234567890.
Use it in the terminal to finalize your Mantil registration:

mantil user activate 1234567890

The Mantil Team
`
	require.Equal(t, expected, content)
}
