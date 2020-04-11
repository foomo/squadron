package configurd

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServiceDeployment_Upgrade(t *testing.T) {
	os.Chdir("example")
	cnf, err := New(".")
	require.NoError(t, err)

	output, err := cnf.Deploy("local", "hello-deployment", "master")
	assert.NoError(t, err)
	t.Log(output)
}
