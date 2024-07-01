package helm

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	chartAPIVersionV2 = "v2"
	defaultChartType  = "application" // application or library
	chartFile         = "Chart.yaml"
	valuesFile        = "values.yaml"
)

func loadChart(path string) (*Chart, error) {
	c := Chart{}
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while opening file: %v", err)
	}
	if err := yaml.Unmarshal(file, &c); err != nil {
		return nil, fmt.Errorf("error while unmarshalling template file: %s", err)
	}
	return &c, nil
}
