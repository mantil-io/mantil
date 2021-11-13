package secret

import (
	_ "embed"
	"time"

	"github.com/mantil-io/mantil/kit/token"
	"github.com/mantil-io/mantil/registration"
)

//go:embed private_key
var privateKey string

func Encode(ut registration.UserToken) (string, error) {
	return token.JWT(privateKey, ut, time.Hour*24*365)
}
