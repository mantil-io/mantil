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

	rec := tf.AsRecord()
	require.Equal(t, "ianic", rec.Name)
	require.Equal(t, "an_account@example.com", rec.Email)
	require.Equal(t, "PM", rec.Position)
	require.Equal(t, "71+", rec.OrganizationSize)
	require.Equal(t, false, rec.Developer)
	require.True(t, rec.CreatedAt > 0)
	require.Len(t, rec.ID, 22)
}
