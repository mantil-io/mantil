package test

import (
	"net/http"
	"testing"

	"github.com/gavv/httpexpect"
	"github.com/mantil-io/mantil/signup"
)

func TestRegister(t *testing.T) {
	api := httpexpect.New(t, apiURL)

	req := signup.RegisterRequest{
		Email: "igor.anic@gmail.com",
	}
	api.POST("/signup/register").
		WithJSON(req).
		Expect().
		Status(http.StatusNoContent)
}
