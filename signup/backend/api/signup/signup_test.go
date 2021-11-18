package signup

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMailBody(t *testing.T) {
	buf, err := ioutil.ReadFile("typeform.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	require.Equal(t, "ianic@manilt.com", tf.Email())
}

func TestTypeform2(t *testing.T) {
	buf, err := ioutil.ReadFile("typeform2.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	require.Equal(t, "an_account@example.com", tf.Email())
}
