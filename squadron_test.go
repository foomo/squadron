package squadron_test

import (
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
		true,
	)
}

func TestConfigTemplateSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-template", "squadron.yaml"),
		},
		path.Join("testdata", "config-template", "squadron.yaml.snapshot"),
		true,
	)
}

func TestConfigNoRenderSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-no-render", "squadron.yaml"),
		},
		path.Join("testdata", "config-no-render", "squadron.yaml.snapshot"),
		false,
	)
}

func testConfigSnapshot(t *testing.T, configs []string, snapshot string, render bool) {
	var cwd string
	testutils.Must(t, util.ValidatePath(".", &cwd))

	sq := squadron.New(cwd, "", configs)

	testutils.Must(t, sq.MergeConfigFiles(), "failed to merge files")

	if render {
		testutils.Must(t, sq.RenderConfig(), "failed to render config")
	}

	yaml, err := sq.GetConfigYAML()
	testutils.Must(t, err, "failed to parse config")

	testutils.MustCheckSnapshot(t, snapshot, yaml)
}
