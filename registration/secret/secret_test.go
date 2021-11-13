package secret_test

import (
	"testing"

	"github.com/mantil-io/mantil/registration"
	"github.com/mantil-io/mantil/registration/secret"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	ut := registration.UserToken{
		ID:        "userid",
		Email:     "developer@mantil.com",
		MachineID: "developerMachine",
	}

	tkn, err := secret.Encode(ut)
	t.Logf("token: %s", tkn)
	require.NoError(t, err)

	ut2, err := registration.Decode(tkn)
	require.NoError(t, err)
	require.Equal(t, ut.ID, ut2.ID)
	require.Equal(t, ut.Email, ut2.Email)
	require.Equal(t, ut.MachineID, ut2.MachineID)
}
