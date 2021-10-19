package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	schemagen "github.com/alecthomas/jsonschema"
	"github.com/mantil-io/mantil/cli/log"

	"github.com/pkg/errors"
	"github.com/qri-io/jsonschema"
	"gopkg.in/yaml.v2"
)

type Schema struct {
	schema *jsonschema.Schema
}

// From initializes a schema from the given struct
func From(v interface{}) (*Schema, error) {
	definition, err := generateSchema(v)
	if err != nil {
		return nil, log.Wrap(err)
	}
	schema, err := initSchema(definition)
	if err != nil {
		return nil, log.Wrap(err)
	}
	return &Schema{
		schema: schema,
	}, nil
}

func generateSchema(v interface{}) ([]byte, error) {
	r := schemagen.Reflector{
		RequiredFromJSONSchemaTags: true,
	}
	schema := r.Reflect(v)
	buf, err := schema.MarshalJSON()
	if err != nil {
		return nil, log.Wrap(err)
	}
	return buf, nil
}

func initSchema(definition []byte) (*jsonschema.Schema, error) {
	jsonschema.RegisterKeyword("definitions", jsonschema.NewDefs)
	jsonschema.LoadDraft2019_09()
	schema := &jsonschema.Schema{}
	if err := json.Unmarshal(definition, schema); err != nil {
		return nil, log.Wrap(err)
	}
	return schema, nil
}

// ValidateYAML validates a yaml file against the schema
func (s *Schema) ValidateYAML(buf []byte) error {
	if len(buf) == 0 {
		return nil
	}
	var m interface{}
	err := yaml.Unmarshal(buf, &m)
	if err != nil {
		return log.Wrap(err)
	}
	m, err = toStringKeys(m, "")
	if err != nil {
		return log.Wrap(err)
	}
	state := s.schema.Validate(context.Background(), m)
	if !state.IsValid() {
		return log.Wrap(formatErrors(*state.Errs))
	}
	return nil
}

func formatErrors(errs []jsonschema.KeyError) error {
	var msgs []string
	for _, e := range errs {
		msg := fmt.Sprintf("%s: %s", e.PropertyPath, e.Message)
		msgs = append(msgs, msg)
	}
	return fmt.Errorf(strings.Join(msgs, "\n"))
}

func toStringKeys(value interface{}, keyPrefix string) (interface{}, error) {
	if mapping, ok := value.(map[interface{}]interface{}); ok {
		dict := make(map[string]interface{})
		for key, entry := range mapping {
			str, ok := key.(string)
			if !ok {
				return nil, formatInvalidKeyError(keyPrefix, key)
			}
			var newKeyPrefix string
			if keyPrefix == "" {
				newKeyPrefix = str
			} else {
				newKeyPrefix = fmt.Sprintf("%s.%s", keyPrefix, str)
			}
			convertedEntry, err := toStringKeys(entry, newKeyPrefix)
			if err != nil {
				return nil, err
			}
			dict[str] = convertedEntry
		}
		return dict, nil
	}
	if list, ok := value.([]interface{}); ok {
		var convertedList []interface{}
		for index, entry := range list {
			newKeyPrefix := fmt.Sprintf("%s[%d]", keyPrefix, index)
			convertedEntry, err := toStringKeys(entry, newKeyPrefix)
			if err != nil {
				return nil, err
			}
			convertedList = append(convertedList, convertedEntry)
		}
		return convertedList, nil
	}
	return value, nil
}

func formatInvalidKeyError(keyPrefix string, key interface{}) error {
	var location string
	if keyPrefix == "" {
		location = "at top level"
	} else {
		location = fmt.Sprintf("in %s", keyPrefix)
	}
	return errors.Errorf("Non-string key %s: %#v", location, key)
}
