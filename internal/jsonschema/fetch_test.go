package jsonschema_test

import (
	"context"
	"testing"

	"github.com/foomo/squadron/internal/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_fetchJSONSchema(t *testing.T) {
	actual, err := jsonschema.Fetch(context.TODO(), "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json")
	require.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, "https://github.com/foomo/squadron/internal/config/config", actual["$id"])
}
