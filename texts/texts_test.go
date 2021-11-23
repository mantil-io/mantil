package texts

import (
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
