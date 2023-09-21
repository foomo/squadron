package helm

import (
	"context"
	"fmt"
	"path"

	"github.com/foomo/squadron/internal/template"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Dependency struct {
	Name       string `yaml:"name,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Alias      string `yaml:"alias,omitempty"`
}

func (cd *Dependency) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!map":
		type wrapper Dependency
		return value.Decode((*wrapper)(cd))
	case "!!str":
		var vString string
		if err := value.Decode(&vString); err != nil {
			return err
		}
		vBytes, err := template.ExecuteFileTemplate(context.Background(), vString, nil, true)
		if err != nil {
			return errors.Wrap(err, "failed to render chart string")
		}
		localChart, err := loadChart(path.Join(string(vBytes), chartFile))
		if err != nil {
			return fmt.Errorf("failed to load local chart: " + vString)
		}
		cd.Name = localChart.Name
		cd.Repository = fmt.Sprintf("file://%v", vString)
		cd.Version = localChart.Version
		return nil
	default:
		return fmt.Errorf("unsupported node tag type for %T: %q", cd, value.Tag)
	}
}
