package helm

import (
	"os"

	"github.com/pkg/errors"
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
		return nil, errors.Wrap(err, "error while opening file")
	}
	if err := yaml.Unmarshal(file, &c); err != nil {
		return nil, errors.Wrap(err, "error while unmarshalling template file")
	}
	return &c, nil
}
