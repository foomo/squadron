package squadron

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
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
	l         *logrus.Entry
	c         Configuration
}

func New(l *logrus.Entry, basePath, namespace string, files []string) (*Squadron, error) {
	sq := Squadron{
		l:         l,
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
	// cleanup old files
	if err := sq.cleanupOutput(sq.chartPath()); err != nil {
		return err
	}
	// generate Chart.yaml and values.yaml
	if err := sq.generateChart(units, sq.chartPath(), sq.name, sq.c.Version); err != nil {
		return err
	}
	// run helm dependency upgrade
	cmd := util.NewHelmCommand(sq.l)
	_, err := cmd.UpdateDependency(sq.name, sq.chartPath())
	if err != nil {
		return err
	}
	return nil
}

func (sq Squadron) Package() error {
	cmd := util.NewHelmCommand(sq.l)
	_, err := cmd.Package(sq.name, sq.chartPath(), sq.basePath)
	return err
}

func (sq Squadron) Down(helmArgs []string) error {
	cmd := util.NewHelmCommand(sq.l)
	cmd.Args("uninstall", sq.name)
	cmd.Args("--namespace", sq.namespace)
	// use extra args
	cmd.Args(helmArgs...)
	// run
	_, err := cmd.Run()
	return err
}

func (sq Squadron) Up(helmArgs []string) error {
	cmd := util.NewHelmCommand(sq.l)
	cmd.Args("upgrade", sq.name, sq.chartPath(), "--install")
	cmd.Args("--namespace", sq.namespace)
	// use extra args
	cmd.Args(helmArgs...)
	// run
	_, err := cmd.Run()
	return err
}

func (sq Squadron) Template(helmArgs []string) (string, error) {
	cmd := util.NewHelmCommand(sq.l)
	cmd.Args("template", sq.name, sq.chartPath())
	cmd.Args("--namespace", sq.namespace)
	// use extra args
	cmd.Args(helmArgs...)
	// run
	return cmd.Run()
}

func (sq Squadron) chartPath() string {
	return path.Join(sq.basePath, defaultOutputDir, sq.name)
}

func (sq Squadron) cleanupOutput(chartPath string) error {
	if _, err := os.Stat(chartPath); err == nil {
		sq.l.Infof("removing dir: %q", chartPath)
		if err := os.RemoveAll(chartPath); err != nil {
			sq.l.Warnf("could not delete chart output directory: %q", err)
		}
	}

	sq.l.Printf("creating dir: %q", chartPath)
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(chartPath, 0744); err != nil {
			return fmt.Errorf("could not create chart output directory: %w", err)
		}
	}
	return nil
}

func (sq Squadron) generateChart(units map[string]Unit, chartPath, chartName, version string) error {
	sq.l.Printf("generating chart %q files in %q", chartName, chartPath)
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
