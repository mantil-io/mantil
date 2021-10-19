package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type schemaExample struct {
	StringField  string               `yaml:"string_field"`
	IntField     int                  `yaml:"int_field"`
	NestedStruct *schemaExampleNested `yaml:"nested_struct"`
}

type schemaExampleNested struct {
	StringField string `yaml:"string_field"`
	IntField    int    `yaml:"int_field"`
}

func TestValidateYAML(t *testing.T) {
	type testcase struct {
		input   string
		isValid bool
	}

	cases := []testcase{
		{
			input: `
string_field: a
int_field: 1
nested_struct:
  string_field: b
  int_field: 2
`,
			isValid: true,
		},
		{
			input: `
string_field: a
int_field: b
nested_struct:
  string_field: b
  int_field: 2
`,
			isValid: false,
		},
		{
			input: `
string_field: a
int_field: 1
extra_field: oh no
nested_struct:
  string_field: b
  int_field: 2
`,
			isValid: false,
		},
		{
			input: `
string_field: a
int_field: 1
nested_struct:
  string_field: b
  int_field: 2
  extra_field: oh no
`,
			isValid: false,
		},
		{
			input: `
string_field: a
int_field: 1
nested_struct:
  string_field: b
  int_field: b
  extra_field: oh no
`,
			isValid: false,
		},
		{
			input: `
string_field: a
nested_struct:
  string_field: b
`,
			isValid: true,
		},
	}

	s, err := From(&schemaExample{})
	assert.Nil(t, err)

	for _, c := range cases {
		err := s.ValidateYAML([]byte(c.input))
		if c.isValid {
			assert.Nil(t, err)
		} else {
			assert.NotNil(t, err)
		}
	}
}
