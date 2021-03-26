package squadron_test

import (
	"path"
	"testing"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/tests/utils"
	"github.com/foomo/squadron/util"
)

func TestConfigSimpleSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-simple", "squadron.yaml"),
		},
		path.Join("testdata", "config-simple", "squadron.yaml.snapshot"),
	)
}

func TestConfigOverrideSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-override", "squadron.yaml"),
			path.Join("testdata", "config-override", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-override", "squadron.yaml.snapshot"),
	)
}

func TestConfigGlobalSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-global", "squadron.yaml"),
			path.Join("testdata", "config-global", "squadron.override.yaml"),
		},
		path.Join("testdata", "config-global", "squadron.yaml.snapshot"),
	)
}

func TestConfigTemplateSnapshot(t *testing.T) {
	testConfigSnapshot(t,
		[]string{
			path.Join("testdata", "config-template", "squadron.yaml"),
		},
		path.Join("testdata", "config-template", "squadron.yaml.snapshot"),
	)
}

func testConfigSnapshot(t *testing.T, configs []string, snapshot string) {
	var cwd string
	testutils.Must(t, util.ValidatePath(".", &cwd))

	sq, err := squadron.New(testutils.Log(), cwd, "", configs)
	testutils.Must(t, err, "failed to init squadron")

	cf, err := sq.Config()
	testutils.Must(t, err, "failed to parse config")

	testutils.MustCheckSnapshot(t, snapshot, []byte(cf))
}
