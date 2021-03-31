package squadron

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
	"github.com/sirupsen/logrus"
)

const (
	defaultOutputDir  = ".squadron"
	chartApiVersionV2 = "v2"
	defaultChartType  = "application" // application or library
	chartFile         = "Chart.yaml"
	valuesFile        = "values.yaml"
)

type Configuration struct {
	Name    string                 `yaml:"name,omitempty"`
	Version string                 `yaml:"version,omitempty"`
	Prefix  string                 `yaml:"prefix,omitempty"`
	Global  map[string]interface{} `yaml:"global,omitempty"`
	Units   map[string]Unit        `yaml:"squadron,omitempty"`
}

type Squadron struct {
	name      string
	basePath  string
	namespace string
	c         Configuration
}

func New(basePath, namespace string, files []string) (*Squadron, error) {
	sq := Squadron{
		basePath:  basePath,
		namespace: namespace,
		c:         Configuration{},
	}

	tv := TemplateVars{}
	if err := mergeSquadronFiles(files, &sq.c, tv); err != nil {
		return nil, err
	}

	sq.name = filepath.Base(basePath)
	if sq.c.Name != "" {
		sq.name = sq.c.Name
	}
	return &sq, nil
}

func (sq Squadron) GetUnits() map[string]Unit {
	return sq.c.Units
}

func (sq Squadron) GetGlobal() map[string]interface{} {
	return sq.c.Global
}

func (sq Squadron) GetConfigYAML() ([]byte, error) {
	return yaml.Marshal(sq.c)
}

func (sq Squadron) Generate(units map[string]Unit) error {
	logrus.Infof("recreating chart output dir %q", sq.chartPath())
	if err := sq.cleanupOutput(sq.chartPath()); err != nil {
		return err
	}
	logrus.Infof("generating chart %q files in %q", sq.name, sq.chartPath())
	if err := sq.generateChart(units, sq.chartPath(), sq.name, sq.c.Version); err != nil {
		return err
	}
	logrus.Infof("running helm dependency update for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().UpdateDependency(sq.name, sq.chartPath())
	return err
}

func (sq Squadron) Package() error {
	logrus.Infof("running helm package for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().Package(sq.name, sq.chartPath(), sq.basePath)
	return err
}

func (sq Squadron) Down(helmArgs []string) error {
	logrus.Infof("running helm uninstall for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().Args("uninstall", sq.name).
		Args("--namespace", sq.namespace).Args(helmArgs...).Run()
	return err
}

func (sq Squadron) Up(helmArgs []string) error {
	logrus.Infof("running helm install for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().
		Args("upgrade", sq.name, sq.chartPath(), "--install").
		Args("--namespace", sq.namespace).Args(helmArgs...).Run()
	return err
}

func (sq Squadron) Template(helmArgs []string) (string, error) {
	logrus.Infof("running helm template for chart: %v", sq.chartPath())
	return util.NewHelmCommand().Args("template", sq.name, sq.chartPath()).
		Args("--namespace", sq.namespace).Args(helmArgs...).Run()
}

func (sq Squadron) chartPath() string {
	return path.Join(sq.basePath, defaultOutputDir, sq.name)
}

func (sq Squadron) cleanupOutput(chartPath string) error {
	if _, err := os.Stat(chartPath); err == nil {
		if err := os.RemoveAll(chartPath); err != nil {
			logrus.Warnf("could not delete chart output directory: %q", err)
		}
	}
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(chartPath, 0744); err != nil {
			return fmt.Errorf("could not create chart output directory: %w", err)
		}
	}
	return nil
}

func (sq Squadron) generateChart(units map[string]Unit, chartPath, chartName, version string) error {
	chart := newChart(chartName, version)
	values := map[string]interface{}{}
	if sq.GetGlobal() != nil {
		values["global"] = sq.GetGlobal()
	}
	for name, unit := range units {
		chart.addDependency(name, unit.Chart)
		values[name] = unit.Values
	}
	if err := chart.generate(chartPath, values); err != nil {
		return err
	}
	return nil
}
