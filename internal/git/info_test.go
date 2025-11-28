package git_test

import (
	"testing"

	"github.com/foomo/squadron/internal/git"
	"github.com/stretchr/testify/require"
)

func TestGetInfo(t *testing.T) {
	t.Setenv("PROJECT_ROOT", "../../")

	info, err := git.GetInfo(t.Context())
	require.NoError(t, err)

	t.Log(info)
}
