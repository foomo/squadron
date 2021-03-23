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

type Build struct {
	Image      string   `yaml:"image"`
	Tag        string   `yaml:"tag"`
	Context    string   `yaml:"context"`
	Dockerfile string   `yaml:"dockerfile"`
	Args       []string `yaml:"args"`
	Labels     []string `yaml:"labels"`
	CacheFrom  []string `yaml:"cache_from"`
	Network    string   `yaml:"network"`
	Target     string   `yaml:"target"`
	ShmSize    string   `yaml:"shm_size"`
	ExtraHosts []string `yaml:"extra_hosts"`
	Isolation  string   `yaml:"isolation"`
}

type Unit struct {
	Chart  interface{}            `yaml:"chart,omitempty"`
	Builds map[string]interface{} `yaml:"builds,omitempty"`
	Values interface{}            `yaml:"values,omitempty"`
}

func (u Unit) getChartDependency() (*ChartDependency, error) {
	if u.Chart == nil {
		return nil, nil
	}
	cd := ChartDependency{}
	if vString, ok := u.Chart.(string); ok {
		localChart, err := loadChart(path.Join(vString, chartFile))
		if err != nil {
			return nil, fmt.Errorf("unable to load chart path %q, error: %v", vString, err)
		}
		cd.Name = localChart.Name
		cd.Repository = fmt.Sprintf("file://%v", vString)
		cd.Version = localChart.Version
	} else {
		out, err := yaml.Marshal(u.Chart)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(out, &cd); err != nil {
			return nil, err
		}
	}
	return &cd, nil
}

func (u Unit) getBuilds() (*Build, error) {
	if len(u.Builds) == 0 {
		return nil, nil
	}
	b := Build{}
	for _, build := range u.Builds {
		if vString, ok := build.(string); ok {
			b.Context = vString
		} else {
			out, err := yaml.Marshal(build)
			if err != nil {
				return nil, err
			}
			if err := yaml.Unmarshal(out, &b); err != nil {
				return nil, err
			}
		}
	}
	return &b, nil
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
	c         *Configuration
	tv        TemplateVars
}

func New(l *logrus.Entry, basePath, namespace string) (*Squadron, error) {
	sq := Squadron{
		l:         l,
		helmCmd:   util.NewHelmCommand(l),
		dockerCmd: util.NewDockerCommand(l),
		basePath:  basePath,
		tv:        TemplateVars{"PWD": basePath, "NS": namespace},
	}
	sq.helmCmd.Args("-n", namespace)

	cFile := configName + defaultYamlExt
	executeSquadronTemplate(sq.c, cFile)

	sq.name = filepath.Base(basePath)
	if sq.c.Name != "" {
		sq.name = sq.c.Name
	}
	return &sq, nil
}

func (sq Squadron) Units() map[string]Unit {
	return sq.c.Units
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
	b, err := u.getBuilds()
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
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
	_, err = sq.dockerCmd.Build(b.Context)
	return err
}

func (sq Squadron) Push(u Unit) error {
	b, err := u.getBuilds()
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	_, err = sq.dockerCmd.Push(b.Image, b.Tag)
	return err
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
		cd, err := unit.getChartDependency()
		if err != nil {
			return err
		}
		chart.addDependency(name, *cd)
		overrides[name] = unit.Values
	}
	if err := chart.generate(chartPath, overrides); err != nil {
		return err
	}
	return nil
}
