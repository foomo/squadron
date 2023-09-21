package squadron_test

import (
	"context"
	"os"
	"path"
	"testing"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/testutils"
	"github.com/foomo/squadron/internal/util"
	"github.com/stretchr/testify/require"
)

func TestConfigSimpleSnapshot(t *testing.T) {
	tests := []struct {
		name     string
		files    []string
		squadron string
		units    []string
	}{
		{
			name:  "blank",
			files: []string{"squadron.yaml"},
		},
		{
			name:  "simple",
			files: []string{"squadron.yaml"},
		},
		{
			name:  "override",
			files: []string{"squadron.yaml", "squadron.override.yaml"},
		},
		{
			name:  "global",
			files: []string{"squadron.yaml", "squadron.override.yaml"},
		},
		{
			name:  "template",
			files: []string{"squadron.yaml"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			config(tt, test.name, test.files, test.squadron, test.units)
		})
	}
}

func config(t *testing.T, name string, files []string, squadronName string, unitNames []string) {
	t.Helper()
	var cwd string
	ctx := context.TODO()
	require.NoError(t, util.ValidatePath(".", &cwd))
	require.NoError(t, os.Setenv("PROJECT_ROOT", "."))

	for i, file := range files {
		files[i] = path.Join("testdata", name, file)
	}
	sq := squadron.New(cwd, "default", files)

	{
		require.NoError(t, sq.MergeConfigFiles(), "failed to merge files")
	}

	{
		require.NoError(t, sq.FilterConfig(squadronName, unitNames), "failed to filter config")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-config-norender.yaml"), sq.ConfigYAML())
	}

	{
		require.NoError(t, sq.RenderConfig(ctx), "failed to render config")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-config.yaml"), sq.ConfigYAML())
	}

	{
		require.NoError(t, sq.RenderConfig(ctx), "failed to render config")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-config.yaml"), sq.ConfigYAML())
	}

	{
		out, err := sq.Template(ctx, nil)
		require.NoError(t, err, "failed to render template")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-template.yaml"), out)
	}
}
