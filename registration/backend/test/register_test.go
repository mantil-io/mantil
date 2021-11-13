package test

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/mantil-io/mantil/registration/backend/api/register"
)

func TestRegister(t *testing.T) {
	api := httpexpect.New(t, apiURL)

	req := register.DefaultRequest{
		// TODO add attributes
	}
	api.POST("/register").
		WithJSON(req).
		Expect().
		ContentType("application/json").
		Status(http.StatusOK).
		JSON().Object().
		Value("TODO")

}
