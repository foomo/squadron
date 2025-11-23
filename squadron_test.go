package squadron_test

import (
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/testutils"
	"github.com/foomo/squadron/internal/util"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/require"
)

func TestConfigSimpleSnapshot(t *testing.T) {
	testingx.Tags(t, tagx.Short)

	pterm.EnableDebugMessages()
	t.Setenv("PROJECT_ROOT", ".")

	tests := []struct {
		name     string
		files    []string
		squadron string
		units    []string
		tags     []string
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
			name:  "extends",
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
			name:  "vars",
			files: []string{"squadron.yaml", "squadron.override.yaml"},
		},
		{
			name:  "template",
			files: []string{"squadron.yaml"},
		},
		{
			name:  "tags",
			tags:  []string{"backend", "-skip"},
			files: []string{"squadron.yaml"},
		},
		{
			name:  "bake",
			files: []string{"squadron.yaml"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			runTestConfig(tt, test.name, test.files, test.squadron, test.units, test.tags)
		})
	}
}

func runTestConfig(t *testing.T, name string, files []string, squadronName string, unitNames, tags []string) {
	t.Helper()

	var cwd string

	ctx := t.Context()
	require.NoError(t, util.ValidatePath(".", &cwd))

	for i, file := range files {
		files[i] = path.Join("testdata", name, file)
	}

	sq := squadron.New(cwd, "default", files)

	t.Run("merge", func(tt *testing.T) {
		require.NoError(t, sq.MergeConfigFiles(ctx), "failed to merge files")
	})

	t.Run("filter", func(tt *testing.T) {
		require.NoError(t, sq.FilterConfig(ctx, squadronName, unitNames, tags), "failed to filter config")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-config-norender.yaml"), sq.ConfigYAML())
	})

	t.Run("render", func(tt *testing.T) {
		require.NoError(t, sq.RenderConfig(ctx), "failed to render config")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-config.yaml"), sq.ConfigYAML())
	})

	t.Run("rerender", func(tt *testing.T) {
		require.NoError(t, sq.RenderConfig(ctx), "failed to render config")
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-config.yaml"), sq.ConfigYAML())
	})

	t.Run("template", func(tt *testing.T) {
		out, err := sq.Template(ctx, nil, 1)
		require.NoError(t, err)
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-template.yaml"), out)
	})

	t.Run("bakefile", func(tt *testing.T) {
		bakefile, err := sq.Bakefile(ctx)
		require.NoError(t, err)
		if len(bakefile.Targets) == 0 {
			return
		}
		for _, target := range bakefile.Targets {
			target.Labels["org.opencontainers.image.version"] = "-"
			target.Labels["org.opencontainers.image.revision"] = "-"
			target.Labels["org.opencontainers.image.created"] = "-"
		}
		out, err := bakefile.HCL()
		require.NoError(t, err)
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-bakefile.yaml"), string(out))
	})
}
