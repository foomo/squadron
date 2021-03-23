package squadron

import (
	"fmt"
	"io/ioutil"
	"path"

	"github.com/foomo/squadron/util"
	"gopkg.in/yaml.v3"
)

type ChartDependency struct {
	Name       string `yaml:"name,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Alias      string `yaml:"alias,omitempty"`
}

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

func (c *Chart) addDependency(alias string, cd ChartDependency) {
	cd.Alias = alias
	c.Dependencies = append(c.Dependencies, cd)
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
