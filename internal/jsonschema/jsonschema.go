package jsonschema

import (
	"context"
	"encoding/json"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// JSONSchema represents the structure of a JSON schema
type JSONSchema struct {
	baseSchema map[string]any
}

// New takes a URL to the base JSON schema and returns a JSONSchema instance
func New() *JSONSchema {
	return &JSONSchema{}
}

func (js *JSONSchema) LoadBaseSchema(ctx context.Context, url string) error {
	baseSchema, err := LoadMap(ctx, url)
	if err != nil {
		return err
	}
	js.baseSchema = baseSchema
	return nil
}

// SetSquadronUnitSchema overrides the base schema at the given path with another JSON schema from a URL
func (js *JSONSchema) SetSquadronUnitSchema(ctx context.Context, squardon, unit, url string) error {
	var ref string
	if strings.HasPrefix(url, "http") {
		ref = strings.TrimPrefix(url, "https:")
		ref = strings.TrimPrefix(ref, "http:")
		ref = strings.TrimPrefix(ref, "//")
	} else {
		ref = path.Clean(url)
		ref = strings.TrimPrefix(ref, "..")
		ref = strings.TrimPrefix(ref, ".")
		ref = strings.TrimPrefix(ref, "/")
	}
	ref = strings.TrimSuffix(ref, "/")
	ref = strings.ReplaceAll(ref, "/", "-")
	ref = strings.ToLower(ref)

	// retrieve definitions
	defsMap := js.ensure(js.baseSchema, "$defs", map[string]any{})

	// add definition
	if _, ok := defsMap[ref]; !ok {
		valuesMap, err := LoadMap(ctx, url)
		if err != nil {
			return errors.Wrap(err, "failed to load map: "+url)
		}
		delete(valuesMap, "$schema")
		js.ensure(defsMap, ref, valuesMap)
	}

	// extend Config
	configMap := js.ensure(defsMap, "Config", map[string]any{})
	configPropertiesMap := js.ensure(configMap, "properties", map[string]any{})
	squadronsMap := js.ensure(configPropertiesMap, "squadron", map[string]any{})
	squadronsPropertiesMap := js.ensure(squadronsMap, "properties", map[string]any{})
	squadronMap := js.ensure(squadronsPropertiesMap, squardon, map[string]any{
		"additionalProperties": map[string]any{
			"$ref": "#/$defs/Unit",
		},
		"type": "object",
	})
	squadronPropertiesMap := js.ensure(squadronMap, "properties", map[string]any{})
	_ = js.ensure(squadronPropertiesMap, unit, map[string]any{
		"anyOf": []map[string]any{
			{
				"$ref": "#/$defs/Unit",
			},
			{
				"type": "object",
				"properties": map[string]any{
					"values": map[string]any{
						"$ref": "#/$defs/" + ref,
					},
				},
			},
		},
	})

	return nil
}

// String outputs the resulting JSON schema as string
func (js *JSONSchema) String() (string, error) {
	output, err := json.Marshal(js.baseSchema)
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// PrettyString outputs the resulting JSON schema as a formatted string
func (js *JSONSchema) PrettyString() (string, error) {
	output, err := json.MarshalIndent(js.baseSchema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func (js *JSONSchema) ensure(source map[string]any, name string, initial map[string]any) map[string]any {
	ret, ok := source[name].(map[string]any)
	if !ok {
		ret = initial
		source[name] = ret
	}
	return ret
}
