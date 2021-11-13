package secret_test

import (
	"testing"

	"github.com/mantil-io/mantil/signup"
	"github.com/mantil-io/mantil/signup/secret"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	ut := signup.UserToken{
		ID:        "userid",
		Email:     "developer@mantil.com",
		MachineID: "developerMachine",
	}

	tkn, err := secret.Encode(ut)
	t.Logf("token: %s", tkn)
	require.NoError(t, err)

	ut2, err := signup.Decode(tkn)
	require.NoError(t, err)
	require.Equal(t, ut.ID, ut2.ID)
	require.Equal(t, ut.Email, ut2.Email)
	require.Equal(t, ut.MachineID, ut2.MachineID)
}
