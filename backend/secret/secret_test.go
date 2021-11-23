package secret_test

import (
	"testing"

	"github.com/mantil-io/mantil/backend/secret"
	cliSecret "github.com/mantil-io/mantil/cli/secret"
	"github.com/mantil-io/mantil/domain/signup"
	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	ut := signup.TokenClaims{
		ActivationCode: "userid",
		ActivationID:   "activationID",
		MachineID:      "developerMachine",
	}

	tkn, err := secret.Encode(ut)
	t.Logf("token: %s", tkn)
	require.NoError(t, err)

	ut2, err := signup.Decode(tkn, cliSecret.SignupPublicKey)
	require.NoError(t, err)
	require.Equal(t, ut.ActivationCode, ut2.ActivationCode)
	require.Equal(t, ut.ActivationID, ut2.ActivationID)
	require.Equal(t, ut.MachineID, ut2.MachineID)
}
