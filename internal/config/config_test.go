package config_test

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	"github.com/foomo/squadron/internal/config"
	"github.com/invopop/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	schema := jsonschema.Reflect(&config.Config{})
	out, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	cwd, err := os.Getwd()
	require.NoError(t, err)
	filename := path.Clean(path.Join(cwd, "..", "..", "squadron.schema.json"))

	expected, err := os.ReadFile(filename)
	require.NoError(t, err)

	if !assert.Equal(t, string(expected), string(out)) {
		require.NoError(t, os.WriteFile(filename, out, 0600))
	}
}
