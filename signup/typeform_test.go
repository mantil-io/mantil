package signup

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

// https://developer.typeform.com/webhooks/example-payload/
// https://mholt.github.io/json-to-go/
func TestTypeformWithExamplePayload(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/typeform.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	require.Equal(t, "ianic@mantil.com", tf.Email())
}

func TestTypeformWithOurSignupFormPayload(t *testing.T) {
	buf, err := ioutil.ReadFile("testdata/typeform_our_form.json")
	require.NoError(t, err)
	var tf TypeformWebhook
	err = json.Unmarshal(buf, &tf)
	require.NoError(t, err)

	require.Equal(t, "an_account@example.com", tf.Email())
}
