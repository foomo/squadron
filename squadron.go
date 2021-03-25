package squadron

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/foomo/squadron/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultOutputDir  = ".output"
	chartApiVersionV2 = "v2"
	defaultChartType  = "application" // application or library
	chartFile         = "Chart.yaml"
	valuesFile        = "values.yaml"
	defaultYamlExt    = ".yaml"
	configName        = "squadron"
)

type Unit struct {
	Chart  ChartDependency        `yaml:"chart,omitempty"`
	Builds map[string]Build       `yaml:"builds,omitempty"`
	Values map[string]interface{} `yaml:"values,omitempty"`
}

type Configuration struct {
	Name    string          `yaml:"name,omitempty"`
	Version string          `yaml:"version,omitempty"`
	Prefix  string          `yaml:"prefix,omitempty"`
	Units   map[string]Unit `yaml:"squadron,omitempty"`
}

type Squadron struct {
	name      string
	l         *logrus.Entry
	helmCmd   *util.HelmCmd
	dockerCmd *util.DockerCmd
	basePath  string
	c         Configuration
}

func New(l *logrus.Entry, basePath, namespace string, files []string) (*Squadron, error) {
	sq := Squadron{
		l:         l,
		helmCmd:   util.NewHelmCommand(l),
		dockerCmd: util.NewDockerCommand(l),
		basePath:  basePath,
		c:         Configuration{},
	}
	sq.helmCmd.Args("-n", namespace)

	tv := TemplateVars{}
	tv.add("PWD", basePath)
	tv.add("NS", namespace)
	cFile := path.Join(basePath, configName+defaultYamlExt)
	if err := executeSquadronTemplate(cFile, &sq.c, tv); err != nil {
		return nil, err
	}
	if err := mergeSquadronFiles(files, &sq.c, tv); err != nil {
		return nil, err
	}

	sq.name = filepath.Base(basePath)
	if sq.c.Name != "" {
		sq.name = sq.c.Name
	}
	return &sq, nil
}

func (sq Squadron) Units() map[string]Unit {
	return sq.c.Units
}

func (sq Squadron) Config() error {
	bs, err := yaml.Marshal(sq.c)
	if err != nil {
		return err
	}
	if _, err := os.Stdout.WriteString(string(bs)); err != nil {
		return err
	}
	return nil
}

func (sq Squadron) Generate(units map[string]Unit) error {
	chartPath := path.Join(sq.basePath, defaultOutputDir, "default", sq.name)
	// cleanup old files
	if err := sq.cleanupOutput(chartPath); err != nil {
		return err
	}
	// generate Chart.yaml and values.yaml
	if err := sq.generateChart(units, chartPath, sq.name, sq.c.Version); err != nil {
		return err
	}
	// run helm dependancy upgrade
	_, err := sq.helmCmd.UpdateDependency(sq.name, chartPath)
	if err != nil {
		return err
	}
	// run helm package to basePath
	_, err = sq.helmCmd.Package(sq.name, chartPath, sq.basePath)
	if err != nil {
		return err
	}
	return nil
}

func (sq Squadron) Down(helmArgs []string) error {
	// use extra args
	sq.helmCmd.Args(helmArgs...)
	// run helm upgrade --install
	_, err := sq.helmCmd.Uninstall(sq.name)
	return err
}

func (sq Squadron) Up(units map[string]Unit, namespace string, helmArgs []string) error {
	chartPath := path.Join(sq.basePath, defaultOutputDir, namespace, sq.name)
	// cleanup old files
	if err := sq.cleanupOutput(chartPath); err != nil {
		return err
	}
	// generate Chart.yaml and values.yaml
	if err := sq.generateChart(units, chartPath, sq.name, sq.c.Version); err != nil {
		return err
	}
	// run helm dependancy upgrade
	_, err := sq.helmCmd.UpdateDependency(sq.name, chartPath)
	if err != nil {
		return err
	}
	// use extra args
	sq.helmCmd.Args(helmArgs...)
	// run helm upgrade --install
	_, err = sq.helmCmd.Install(sq.name, chartPath)
	return err
}

func (sq Squadron) Build(u Unit) error {
	for _, b := range u.Builds {
		dockerCmd := sq.dockerCmd
		dockerCmd.Option("-t", fmt.Sprintf("%v:%v", b.Image, b.Tag))
		dockerCmd.Option("--file", b.Dockerfile)
		dockerCmd.ListOption("--build-arg", b.Args)
		dockerCmd.ListOption("--label", b.Labels)
		dockerCmd.ListOption("--cache-from", b.CacheFrom)
		dockerCmd.Option("--network", b.Network)
		dockerCmd.Option("--target", b.Target)
		dockerCmd.Option("--shm-size", b.ShmSize)
		dockerCmd.ListOption("--add-host", b.ExtraHosts)
		dockerCmd.Option("--isolation", b.Isolation)
		if _, err := sq.dockerCmd.Build(b.Context); err != nil {
			return err
		}
	}
	return nil
}

func (sq Squadron) Push(u Unit) error {
	for _, b := range u.Builds {
		if _, err := sq.dockerCmd.Push(b.Image, b.Tag); err != nil {
			return err
		}
	}
	return nil
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
	overrides := map[string]interface{}{}
	for name, unit := range units {
		chart.addDependency(name, unit.Chart)
		overrides[name] = unit.Values
	}
	if err := chart.generate(chartPath, overrides); err != nil {
		return err
	}
	return nil
}
