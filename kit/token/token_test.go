package token_test

import (
	"testing"
	"time"

	. "github.com/mantil-io/mantil/kit/token"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	pub, priv, err := KeyPair()
	require.NoError(t, err)
	require.Len(t, pub, 43)
	require.Len(t, priv, 86)

	// fmt.Printf("pub: %s\n", pub)
	// fmt.Printf("priv: %s\n", priv)
}

type Claims struct {
	UID   string `json:"u,omitempty"`
	Email string `json:"e,omitempty"`
}

func TestJWT(t *testing.T) {
	pub, priv, err := KeyPair()
	require.NoError(t, err)

	c := Claims{
		UID:   "4ce6e5a0-451b-453a-9b0b-30e037f6029e",
		Email: "someone@example.com",
	}

	j, err := JWT(priv, c, time.Hour)
	require.NoError(t, err)

	var c2 Claims

	err = Decode(j, pub, &c2)
	require.NoError(t, err)

	require.Equal(t, c, c2)
}
