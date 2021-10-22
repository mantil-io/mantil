package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateName(t *testing.T) {
	type testcase struct {
		input     string
		errorType error
	}

	cases := []testcase{
		{
			input:     "name",
			errorType: nil,
		},
		{
			input:     "NAME",
			errorType: nil,
		},
		{
			input:     "name-123",
			errorType: nil,
		},
		{
			input:     "name_123",
			errorType: nil,
		},
		{
			input:     "some-very-long-name",
			errorType: &ErrNameTooLong{},
		},
		{
			input:     "neko-dugačko-ime",
			errorType: &ErrNameTooLong{},
		},
		{
			input:     "kraće-ime",
			errorType: &ErrForbiddenCharacters{},
		},
	}

	for _, c := range cases {
		err := ValidateName(c.input)
		if c.errorType == nil {
			assert.Nil(t, err)
		} else {
			assert.IsType(t, c.errorType, err)
		}
	}
}
