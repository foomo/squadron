package config

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/foomo/squadron/internal/template"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Chart struct {
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	Repository string `json:"repository,omitempty" yaml:"repository,omitempty"`
	Schema     string `json:"schema,omitempty" yaml:"schema,omitempty"`
	Version    string `json:"version,omitempty" yaml:"version,omitempty"`
	Alias      string `json:"alias,omitempty" yaml:"alias,omitempty"`
}

func (d *Chart) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!map":
		type wrapper Chart
		return value.Decode((*wrapper)(d))
	case "!!str":
		var vString string
		if err := value.Decode(&vString); err != nil {
			return err
		}
		vBytes, err := template.ExecuteFileTemplate(context.Background(), vString, nil, true)
		if err != nil {
			return errors.Wrap(err, "failed to render chart string")
		}
		localChart, err := loadChart(path.Join(string(vBytes), "Chart.yaml"))
		if err != nil {
			return errors.New("failed to load local chart: " + vString)
		}
		d.Name = localChart.Name
		d.Repository = fmt.Sprintf("file://%v", vString)
		d.Version = localChart.Version
		wd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "failed to get working directory")
		}
		schemaPath := string(vBytes)
		if value, err := filepath.Rel(wd, string(vBytes)); err == nil {
			schemaPath = value
		}

		if _, err := os.Stat(path.Join(schemaPath, "values.schema.json")); err == nil {
			d.Schema = path.Join(schemaPath, "values.schema.json")
		}
		return nil
	default:
		return fmt.Errorf("unsupported node tag type for %T: %q", d, value.Tag)
	}
}

func (d *Chart) String() string {
	return fmt.Sprintf("%s/%s:%s", d.Repository, d.Name, d.Version)
}

func loadChart(name string) (*Chart, error) {
	c := Chart{}
	file, err := os.ReadFile(name)
	if err != nil {
		return nil, errors.Wrap(err, "error while opening file")
	}
	if err := yaml.Unmarshal(file, &c); err != nil {
		return nil, errors.Wrap(err, "error while unmarshalling template file")
	}
	return &c, nil
}
