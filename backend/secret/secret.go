// Package secret is separate package because it contains private key.
// That should not be part of the cli.
// Should be only included in backend.
package secret

import (
	_ "embed"
	"time"

	"github.com/mantil-io/mantil/domain/signup"
	"github.com/mantil-io/mantil/kit/token"
)

//go:embed private_key
var privateKey string

func Encode(tc signup.TokenClaims) (string, error) {
	return token.JWT(privateKey, tc, time.Hour*24*365)
}

func TokenForTests(machineID string) string {
	tc := signup.TokenClaims{
		ActivationCode: signup.TestActivationCode,
		Email:          signup.TestEmail,
		MachineID:      machineID,
		CreatedAt:      time.Now().UnixMilli(),
	}
	jwt, _ := Encode(tc)
	return jwt
}
