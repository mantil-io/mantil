package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type schemaExample struct {
	StringField  string               `yaml:"string_field" jsonschema:"minLength=1,maxLength=2"`
	IntField     int                  `yaml:"int_field"`
	NestedStruct *schemaExampleNested `yaml:"nested_struct"`
}

type schemaExampleNested struct {
	StringField string `yaml:"string_field"`
	IntField    int    `yaml:"int_field" jsonschema:"minimum=1,maximum=2"`
}

func TestValidateYAML(t *testing.T) {
	type testcase struct {
		input   string
		isValid bool
	}

	cases := []testcase{
		// empty input
		{
			input:   ``,
			isValid: true,
		},
		{
			input:   `#comment`,
			isValid: true,
		},
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
		// wrong type
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
		// extra field
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
		// missing non-required fields
		{
			input: `
string_field: a
nested_struct:
  string_field: b
`,
			isValid: true,
		},
		// invalid values (min, max length etc.)
		{
			input: `
string_field:
`,
			isValid: false,
		},
		{
			input: `
string_field: abc
`,
			isValid: false,
		},
		{
			input: `
nested_struct:
  int_field: 0
`,
			isValid: false,
		},
		{
			input: `
nested_struct:
  int_field: 3
`,
			isValid: false,
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
