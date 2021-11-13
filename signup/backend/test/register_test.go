package test

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/mantil-io/mantil/domain"
	"github.com/mantil-io/mantil/signup"
	"github.com/stretchr/testify/require"
)

func TestSignup(t *testing.T) {
	api := httpexpect.New(t, apiURL)

	registerRequest := signup.RegisterRequest{
		Name:             "Pero Zdero",
		Email:            signup.TestEmail,
		OrganizationSize: "Only me",
		Position:         "Other",
	}
	api.POST("/signup/register").
		WithJSON(registerRequest).
		Expect().
		Status(http.StatusNoContent)

	machineID := domain.MachineID()
	activateRequest := signup.ActivateRequest{
		ID:        signup.TestID,
		MachineID: machineID,
	}
	jwt := api.POST("/signup/activate").
		WithJSON(activateRequest).
		Expect().
		Status(http.StatusOK).
		Text().Raw()

	tc, err := signup.Decode(jwt)
	require.NoError(t, err)

	t.Logf("jwt: %s", jwt)

	require.Equal(t, machineID, tc.MachineID)
	require.Equal(t, signup.TestEmail, tc.Email)
	require.Equal(t, signup.TestID, tc.ID)
}
