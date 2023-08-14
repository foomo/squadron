package testutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MustWriteSnapshot updates the snapshot file for a given test t.
func MustWriteSnapshot(t *testing.T, name string, content string) {
	t.Helper()
	Must(t, os.WriteFile(name, []byte(content), 0o600), "failed to update snapshot", name)
}

// MustReadSnapshot reads the snapshot file for a given test t.
func MustReadSnapshot(t *testing.T, name string) string {
	t.Helper()
	g, err := os.ReadFile(name)
	Must(t, err, "failed reading file", name)
	return string(g)
}

// MustCheckSnapshot compares v with its snapshot file
func MustCheckSnapshot(t *testing.T, name, yaml string) {
	t.Helper()
	if *UpdateFlag {
		MustWriteSnapshot(t, name, yaml)
	}
	snapshot := MustReadSnapshot(t, name)
	if !assert.YAMLEq(t, snapshot, yaml) {
		t.Fatalf("err: %s not equal to %s", yaml, snapshot)
	}
}
