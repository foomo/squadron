package squadron_test

import (
	"context"
	"path"
	"testing"

	"github.com/foomo/squadron"
	testutils "github.com/foomo/squadron/tests/utils"
	"github.com/foomo/squadron/util"
)

func TestConfigSimpleSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-simple", "squadron.yaml"),
		},
		path.Join("testdata", "config-simple", "squadron.yaml.snapshot"),
		nil,
		true,
	)
}

func TestConfigNoValuesSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-no-values", "squadron.yaml"),
		},
		path.Join("testdata", "config-no-values", "squadron.yaml.snapshot"),
		nil,
		true,
	)
}

func TestConfigOverrideSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-override", "squadron.yaml"),
			path.Join("testdata", "config-override", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-override", "squadron.yaml.snapshot"),
		nil,
		true,
	)
}

func TestConfigGlobalSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-global", "squadron.yaml"),
			path.Join("testdata", "config-global", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-global", "squadron.yaml.snapshot"),
		nil,
		true,
	)
}

func TestConfigTemplateSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-template", "squadron.yaml"),
		},
		path.Join("testdata", "config-template", "squadron.yaml.snapshot"),
		nil,
		true,
	)
}

func TestConfigTemplateFrontendSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-template-frontend", "squadron.yaml"),
		},
		path.Join("testdata", "config-template-frontend", "squadron.yaml.snapshot"),
		[]string{"frontend"},
		true,
	)
}

func TestConfigNoRenderSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-no-render", "squadron.yaml"),
		},
		path.Join("testdata", "config-no-render", "squadron.yaml.snapshot"),
		nil,
		false,
	)
}

func TestConfigOverrideSnapshotNulled(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-override-null", "squadron.yaml"),
			path.Join("testdata", "config-override-null", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-override-null", "squadron.yaml.snapshot"),
		nil,
		true,
	)
}

func testConfigSnapshot(t *testing.T, configs []string, snapshot string, units []string, render bool) {
	var cwd string
	testutils.Must(t, util.ValidatePath(".", &cwd))

	sq := squadron.New(cwd, "", configs)

	testutils.Must(t, sq.MergeConfigFiles(), "failed to merge files")

	if units != nil {
		testutils.Must(t, sq.FilterConfig(units), "failed to filter units")
	}

	if render {
		testutils.Must(t, sq.RenderConfig(context.Background()), "failed to render config")
	}

	testutils.MustCheckSnapshot(t, snapshot, sq.GetConfigYAML())
}
