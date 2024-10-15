package jsonschema

import (
	"context"
	"encoding/json"
	"os"
)

// JSONSchema represents the structure of a JSON schema
type JSONSchema struct {
	filename   string
	baseSchema map[string]any
}

// New takes a URL to the base JSON schema and returns a JSONSchema instance
func New(filename string) *JSONSchema {
	return &JSONSchema{
		filename: filename,
	}
}

func (js *JSONSchema) LoadBaseSchema(ctx context.Context, url string) error {
	baseSchema, err := Fetch(ctx, url)
	if err != nil {
		return err
	}
	js.baseSchema = baseSchema
	return nil
}

// SetSquadronUnitSchema overrides the base schema at the given path with another JSON schema from a URL
func (js *JSONSchema) SetSquadronUnitSchema(ctx context.Context, squardon, unit, url string) error {
	valuesMap, err := Fetch(ctx, url)
	if err != nil {
		return err
	}
	delete(valuesMap, "$schema")

	defsMap := js.ensure(js.baseSchema, "$defs", map[string]any{})
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
		"$ref": "#/$defs/Unit",
		"properties": map[string]any{
			"values": valuesMap,
		},
	})

	return nil
}

func (js *JSONSchema) Write() error {
	out, err := js.String()
	if err != nil {
		return err
	}
	return os.WriteFile(js.filename, []byte(out), 0600)
}

func (js *JSONSchema) WritePretty() error {
	out, err := js.PrettyString()
	if err != nil {
		return err
	}
	return os.WriteFile(js.filename, []byte(out), 0600)
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
