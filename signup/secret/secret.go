// Package secret is separate package because it contains private key.
// That should not be part of the cli.
// Should be only included in backend.
package secret

import (
	_ "embed"
	"time"

	"github.com/mantil-io/mantil/kit/token"
	"github.com/mantil-io/mantil/signup"
)

//go:embed private_key
var privateKey string

func Encode(ut signup.TokenClaims) (string, error) {
	return token.JWT(privateKey, ut, time.Hour*24*365)
}
