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
		s.noEmail = true
		req := signup.RegisterRequest{
			Email:            "hello@mantil.com",
			Name:             "test",
			Position:         "haus majstor",
			OrganizationSize: "1",
		}
		rec, err := s.register(ctx, req)
		require.NoError(t, err)
		require.Len(t, rec.ID, 22)
		require.Len(t, rec.ActivationCode, 22)
		require.True(t, rec.ActivatedAt == 0)
		require.Len(t, rec.Token, 0)
		activationCode = rec.ActivationCode

		t.Logf("registration: %#v", rec)
	})

	t.Run("activate", func(t *testing.T) {
		s := New()
		s.noEmail = true
		req := signup.ActivateRequest{
			ActivationCode: activationCode,
			WorkspaceID:    "my-workspace",
			MachineID:      "my-machine",
		}
		rec, err := s.activate(ctx, req)
		require.NoError(t, err)
		require.True(t, rec.ActivatedAt > 0)
		require.True(t, len(rec.Token) > 50)

		t.Logf("activation: %#v", rec)
	})

	t.Run("typeform register", func(t *testing.T) {
		s := New()
		s.noEmail = true
		buf, err := ioutil.ReadFile("../../../domain/signup/testdata/typeform.json")
		require.NoError(t, err)
		var tf signup.TypeformWebhook
		err = json.Unmarshal(buf, &tf)
		require.NoError(t, err)

		rec, err := s.typeform(ctx, tf)
		require.NoError(t, err)
		require.Len(t, rec.ID, 22)
		require.Len(t, rec.ActivationCode, 22)
		require.True(t, rec.ActivatedAt == 0)
		require.Len(t, rec.Token, 0)
		activationCode = rec.ActivationCode
		t.Logf("typeform registration: %#v", rec)
	})

	t.Run("activate", func(t *testing.T) {
		s := New()
		s.noEmail = true
		req := signup.ActivateRequest{
			ActivationCode: activationCode,
			WorkspaceID:    "my-workspace2",
			MachineID:      "my-machine2",
		}
		rec, err := s.activate(ctx, req)
		require.NoError(t, err)
		require.True(t, rec.ActivatedAt > 0)
		require.True(t, len(rec.Token) > 50)

		t.Logf("activation: %#v", rec)
	})

}
