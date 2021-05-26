package testutils

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

// MustWriteSnapshot updates the snapshot file for a given test t.
func MustWriteSnapshot(t *testing.T, name string, content string) {
	Must(t, ioutil.WriteFile(name, []byte(content), 0600), "failed to update snapshot", name)
}

// MustReadSnapshot reads the snapshot file for a given test t.
func MustReadSnapshot(t *testing.T, name string) string {
	g, err := ioutil.ReadFile(name)
	Must(t, err, "failed reading file", name)
	return string(g)
}

// MustCheckSnapshot compares v with its snapshot file
func MustCheckSnapshot(t *testing.T, name, yaml string) {
	if *UpdateFlag {
		MustWriteSnapshot(t, name, yaml)
	}
	snapshot := MustReadSnapshot(t, name)
	if !assert.YAMLEq(t, string(snapshot), yaml) {
		t.Fatalf("err: %s not equal to %s", yaml, snapshot)
	}
}
