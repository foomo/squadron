package squadron_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/config"
	"github.com/foomo/squadron/internal/testutils"
	"github.com/foomo/squadron/internal/util"
	"github.com/invopop/jsonschema"
	"github.com/pterm/pterm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	t.Parallel()
	testingx.Tags(t, tagx.Short)

	cwd, err := os.Getwd()
	require.NoError(t, err)

	reflector := new(jsonschema.Reflector)
	require.NoError(t, reflector.AddGoComments("github.com/foomo/squadron", "./"))
	schema := reflector.Reflect(&config.Config{})
	actual, err := json.MarshalIndent(schema, "", "  ")
	require.NoError(t, err)

	filename := path.Join(cwd, "squadron.schema.json")
	expected, err := os.ReadFile(filename)
	if !errors.Is(err, os.ErrNotExist) {
		require.NoError(t, err)
	}

	if !assert.Equal(t, string(expected), string(actual)) {
		require.NoError(t, os.WriteFile(filename, actual, 0600))
	}
}

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
	ctx := context.TODO()
	require.NoError(t, util.ValidatePath(".", &cwd))

	for i, file := range files {
		files[i] = path.Join("testdata", name, file)
	}
	sq := squadron.New(cwd, "default", files)

	{
		require.NoError(t, sq.MergeConfigFiles(ctx), "failed to merge files")
	}

	{
		require.NoError(t, sq.FilterConfig(ctx, squadronName, unitNames, tags), "failed to filter config")
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
		out, err := sq.Template(ctx, nil, 1)
		require.NoError(t, err)
		testutils.Snapshot(t, path.Join("testdata", name, "snapshop-template.yaml"), out)
	}
}
