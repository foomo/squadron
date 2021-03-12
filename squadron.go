package squadron

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"github.com/foomo/squadron/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	defaultOutputDir  = ".output"
	chartApiVersionV2 = "v2"
	defaultChartType  = "application" // application or library
	chartFile         = "Chart.yaml"
	valuesFile        = "values.yaml"
)

type Chart struct {
	APIVersion   string            `yaml:"apiVersion"`
	Name         string            `yaml:"name,omitempty"`
	Description  string            `yaml:"description,omitempty"`
	Type         string            `yaml:"type,omitempty"`
	Version      string            `yaml:"version,omitempty"`
	Dependencies []ChartDependency `yaml:"dependencies,omitempty"`
}

func newChart(name, version string) *Chart {
	return &Chart{
		APIVersion:  chartApiVersionV2,
		Name:        name,
		Description: fmt.Sprintf("A helm parent chart for squadron %v", name),
		Type:        defaultChartType,
		Version:     version,
	}
}

func (c *Chart) addDependency(alias string, d interface{}) error {
	cd := ChartDependency{}
	if depString, ok := d.(string); ok {
		localChart, err := loadChart(path.Join(depString, chartFile))
		if err != nil {
			return fmt.Errorf("unable to load chart path %q, error: %v", depString, err)
		}
		cd.Name = localChart.Name
		cd.Repository = depString
		cd.Version = localChart.Version
	} else if depStruct, ok := d.(ChartDependency); ok {
		cd = depStruct
	} else {
		return fmt.Errorf("incorrect format %q for chart field", d)
	}
	cd.Alias = alias
	c.Dependencies = append(c.Dependencies, cd)
	return nil
}

func loadChart(path string) (*Chart, error) {
	c := Chart{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while opening file: %v", err)
	}
	if err := yaml.Unmarshal(file, &c); err != nil {
		return nil, fmt.Errorf("error while unmarshalling template file: %s", err)
	}
	return &c, nil
}

func (c Chart) generate(chartPath string, overrides interface{}) error {
	// generate Chart.yaml
	if err := util.GenerateYaml(path.Join(chartPath, chartFile), c); err != nil {
		return err
	}
	// generate values.yaml
	if err := util.GenerateYaml(path.Join(chartPath, valuesFile), overrides); err != nil {
		return err
	}
	return nil
}

type Build struct {
	Context    string   `yaml:"context,omitempty"`
	Dockerfile string   `yaml:"dockerfile,omitempty"`
	Args       []string `yaml:"args,omitempty"`
	Labels     []string `yaml:"labels,omitempty"`
	CacheFrom  []string `yaml:"cache_from,omitempty"`
	Network    string   `yaml:"network,omitempty"`
	Target     string   `yaml:"target,omitempty"`
	ShmSize    string   `yaml:"shm_size,omitempty"`
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`
	Isolation  string   `yaml:"isolation,omitempty"`
}

type ChartDependency struct {
	Name       string `yaml:"name,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Alias      string `yaml:"alias,omitempty"`
}

type Unit struct {
	Chart  interface{} `yaml:"chart,omitempty"`
	Build  interface{} `yaml:"build,omitempty"`
	Values interface{} `yaml:"values,omitempty"`
}

type Configuration struct {
	name     string
	Version  string          `yaml:"version,omitempty"`
	Prefix   string          `yaml:"prefix,omitempty"`
	Squadron map[string]Unit `yaml:"squadron,omitempty"`
}

type Squadron struct {
	l        *logrus.Entry
	helmCmd  *util.HelmCmd
	basePath string
	c        Configuration
}

func (sq Squadron) Units() map[string]Unit {
	return sq.c.Squadron
}

func New(l *logrus.Entry, basePath, namespace string) (*Squadron, error) {
	sq := Squadron{l: l, helmCmd: util.NewHelmCommand(l), basePath: basePath}
	sq.helmCmd.Args("-n", namespace)
	// todo load and parse configuration file
	return &sq, nil
}

func (sq Squadron) Up(units map[string]Unit, namespace string, helmArgs ...string) error {
	chartPath := path.Join(sq.basePath, defaultOutputDir, namespace, sq.c.name)
	// cleanup old files
	if err := sq.cleanupOutput(chartPath); err != nil {
		return err
	}
	// generate Chart.yaml and values.yaml
	if err := sq.generateChart(units, chartPath, sq.c.name, sq.c.Version); err != nil {
		return err
	}
	log.Println()
	// run helm dependancy upgrade
	_, err := sq.helmCmd.UpdateDependency(sq.c.name, chartPath)
	if err != nil {
		return err
	}
	// run helm upgrade --install --create-namespace
	sq.helmCmd.Args(helmArgs...)
	_, err = sq.helmCmd.Install(sq.c.name, chartPath)
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
	sq.l.Printf("generating chart files in %q", chartName, chartPath)
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
