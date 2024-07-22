package helm

import (
	"fmt"
	"path"

	"github.com/foomo/squadron/internal/util"
)

type Chart struct {
	APIVersion   string       `yaml:"apiVersion"`
	Name         string       `yaml:"name,omitempty"`
	Description  string       `yaml:"description,omitempty"`
	Type         string       `yaml:"type,omitempty"`
	Version      string       `yaml:"version,omitempty"`
	Dependencies []Dependency `yaml:"dependencies,omitempty"`
}

func NewChart(name, version string) *Chart {
	return &Chart{
		APIVersion:  chartAPIVersionV2,
		Name:        name,
		Description: fmt.Sprintf("A helm parent chart for squadron %v", name),
		Type:        defaultChartType,
		Version:     version,
	}
}

func (c *Chart) AddDependency(alias string, cd Dependency) {
	cd.Alias = alias
	c.Dependencies = append(c.Dependencies, cd)
}

func (c *Chart) Generate(chartPath string, overrides any) error {
	// generate Chart.yaml
	if err := util.GenerateYaml(path.Join(chartPath, chartFile), c); err != nil {
		return err
	}
	// generate values.yaml
	return util.GenerateYaml(path.Join(chartPath, valuesFile), overrides)
}
