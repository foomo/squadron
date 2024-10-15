package jsonschema_test

import (
	"fmt"
	"testing"

	"github.com/foomo/squadron/internal/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMerger(t *testing.T) {
	// Example usage
	baseURL := "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json"
	overrideURL := "https://raw.githubusercontent.com/foomo/helm-charts/refs/heads/main/charts/namespace/values.schema.json"

	// Create the JSONSchema object
	js, err := jsonschema.New(baseURL)
	require.NoError(t, err)

	// Override the base schema
	err = js.SetSquadronUnitSchema("site", "namespace", overrideURL)
	require.NoError(t, err)

	// Print the resulting schema
	result, err := js.JSON()
	require.NoError(t, err)

	if !assert.JSONEq(t, "{}", result) {
		fmt.Println(result)
	}
}
