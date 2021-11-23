package signup

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	"github.com/mantil-io/mantil/domain/signup"
	"github.com/stretchr/testify/require"
)

const testsProfileEnv = "MANTIL_TESTS_AWS_PROFILE"

func TestIntegration(t *testing.T) {
	awsProfile, ok := os.LookupEnv(testsProfileEnv)
	if !ok {
		t.Logf("environment vairable %s not found", testsProfileEnv)
		t.Skip()
	}
	t.Setenv("AWS_PROFILE", awsProfile)

	ctx := context.TODO()
	var activationCode string
	t.Run("register", func(t *testing.T) {
		s := New()
		req := signup.RegisterRequest{
			Email:            "hello@mantil.com",
			Name:             "test",
			Position:         "haus majstor",
			OrganizationSize: "1",
		}
		rec, err := s.register(ctx, req)
		require.NoError(t, err)
		require.Len(t, rec.ActivationCode, 22)
		require.True(t, rec.Developer)
		require.Equal(t, rec.Email, req.Email)
		require.True(t, rec.CreatedAt > 0)

		activationCode = rec.ActivationCode

		t.Logf("registration: %#v", rec)
	})
	t.Logf("activationCode: %s", activationCode)

	t.Run("activate", func(t *testing.T) {
		s := New()
		req := signup.ActivateRequest{
			ActivationCode: activationCode,
			WorkspaceID:    "my-workspace",
			MachineID:      "my-machine",
		}
		ar, rr, err := s.activate(ctx, req)
		require.NoError(t, err)

		require.Equal(t, ar.ActivationCode, rr.ActivationCode)
		require.Equal(t, ar.ActivationCode, activationCode)
		require.Equal(t, ar.WorkspaceID, req.WorkspaceID)
		require.Equal(t, ar.MachineID, req.MachineID)
		require.True(t, len(ar.Token) > 50)
		require.Len(t, rr.Activations, 1)
	})

	t.Run("activate with same code", func(t *testing.T) {
		s := New()
		req := signup.ActivateRequest{
			ActivationCode: activationCode,
			WorkspaceID:    "my-workspace-2",
			MachineID:      "my-machine-2",
		}
		ar, rr, err := s.activate(ctx, req)
		require.NoError(t, err)

		require.Equal(t, ar.ActivationCode, rr.ActivationCode)
		require.Equal(t, ar.ActivationCode, activationCode)
		require.Equal(t, ar.WorkspaceID, req.WorkspaceID)
		require.Equal(t, ar.MachineID, req.MachineID)
		require.True(t, len(ar.Token) > 50)
		require.Len(t, rr.Activations, 2)
	})

	t.Run("typeform register", func(t *testing.T) {
		s := New()
		buf, err := ioutil.ReadFile("../../../domain/signup/testdata/typeform.json")
		require.NoError(t, err)
		var tf signup.TypeformWebhook
		err = json.Unmarshal(buf, &tf)
		require.NoError(t, err)

		rec, err := s.typeform(ctx, tf)
		require.NoError(t, err)
		require.Len(t, rec.ActivationCode, 22)
		activationCode = rec.ActivationCode
		t.Logf("typeform registration: %#v", rec)
	})

	t.Run("activate", func(t *testing.T) {
		s := New()
		req := signup.ActivateRequest{
			ActivationCode: activationCode,
			WorkspaceID:    "my-workspace-3",
			MachineID:      "my-machine-3",
		}
		ar, _, err := s.activate(ctx, req)
		require.NoError(t, err)
		require.True(t, len(ar.Token) > 50)
	})
}
