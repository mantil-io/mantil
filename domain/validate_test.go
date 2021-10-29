package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	type testcase struct {
		input   string
		isValid bool
	}

	cases := []testcase{
		{
			input:   "name",
			isValid: true,
		},
		{
			input:   "NAME",
			isValid: true,
		},
		{
			input:   "name-123",
			isValid: true,
		},
		{
			input:   "name_123",
			isValid: true,
		},
		{
			input:   "some-very-long-name",
			isValid: false,
		},
		{
			input:   "neko-dugačko-ime",
			isValid: false,
		},
		{
			input:   "kraće-ime",
			isValid: false,
		},
	}

	for _, c := range cases {
		err := ValidateName(c.input)
		assert.Equal(t, c.isValid, err == nil)
	}
}
