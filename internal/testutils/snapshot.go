package testutils

import (
	"os"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Snapshot compares v with its snapshot file
func Snapshot(t *testing.T, name, yaml string) {
	t.Helper()

	snapshot := readSnapshot(t, name)
	if *UpdateFlag || snapshot == "" {
		writeSnapshot(t, name, yaml)
	}

	assert.YAMLEq(t, snapshot, yaml)
}

// writeSnapshot updates the snapshot file for a given test t.
func writeSnapshot(t *testing.T, name string, content string) {
	t.Helper()
	assert.NoError(t, os.WriteFile(name, []byte(content), 0600), "failed to update snapshot", name)
}

// readSnapshot reads the snapshot file for a given test t.
func readSnapshot(t *testing.T, name string) string {
	t.Helper()

	g, err := os.ReadFile(name)
	if !errors.Is(err, os.ErrNotExist) {
		require.NoError(t, err, "failed reading file", name)
	}

	return string(g)
}
