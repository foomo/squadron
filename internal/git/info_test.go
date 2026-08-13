package git_test

import (
	"testing"

	"github.com/foomo/squadron/internal/git"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	t.Setenv("PROJECT_ROOT", "../../")

	info, err := git.GetInfo(t.Context())
	require.NoError(t, err)

	assert.Equal(t, "https://github.com/foomo/squadron", info.URL)
	assert.NotEmpty(t, info.Commit)
	assert.NotEmpty(t, info.Ref)
}
