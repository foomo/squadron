package config

import (
	"bytes"
	"context"
	"os"
	"path"
	"sort"
	"strings"

	"dario.cat/mergo"
	"github.com/foomo/squadron/internal/template"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
)

type Unit struct {
	// Chart settings
	Chart Chart `json:"chart,omitempty" yaml:"chart,omitempty" jsonschema:"anyof_type=string,anyof_ref=#/$defs/Chart"`
	// Kustomize files path
	Kustomize string `json:"kustomize,omitempty" yaml:"kustomize,omitempty"`
	// List of tags
	Tags Tags `json:"tags,omitempty" yaml:"tags,omitempty"`
	// Installation priority, higher comes first
	Priority int `json:"priority,omitempty" yaml:"priority,omitempty"`
	// Map of containers to build
	Builds map[string]Build `json:"builds,omitempty" yaml:"builds,omitempty"`
	// Chart values
	Values map[string]any `json:"values,omitempty" yaml:"values,omitempty"`
	// Extend chart values
	Extends string `json:"extends,omitempty" yaml:"extends,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (u *Unit) UnmarshalYAML(value *yaml.Node) error {
	type wrapper Unit
	if err := value.Decode((*wrapper)(u)); err != nil {
		return err
	}
	if u.Extends != "" {
		// render filename
		filename, err := template.ExecuteFileTemplate(context.Background(), u.Extends, nil, true)
		if err != nil {
			return errors.Wrap(err, "failed to render defaults filename")
		}

		// read defaults
		defaults, err := os.ReadFile(string(filename))
		if err != nil {
			return errors.Wrap(err, "failed to read defaults")
		}

		var m map[string]any
		if err := yaml.Unmarshal(defaults, &m); err != nil {
			return errors.Wrap(err, "failed to unmarshal defaults")
		}
		if err := mergo.Merge(&m, u.Values, mergo.WithAppendSlice, mergo.WithOverride, mergo.WithSliceDeepCopy); err != nil {
			return err
		}

		u.Extends = ""
		u.Values = m
	}
	return nil
}

// JSONSchemaProperty type workaround
func (Unit) JSONSchemaProperty(prop string) any {
	var x any
	if prop == "chart" {
		return x
	}
	return nil
}

func (u *Unit) ValuesYAML(global map[string]any) ([]byte, error) {
	values := u.Values
	if values == nil {
		values = map[string]any{}
	}
	if global != nil {
		if _, ok := values["global"]; !ok {
			values["global"] = global
		}
	}
	return yamlv2.Marshal(values)
}

func (u *Unit) BuildNames() []string {
	ret := make([]string, 0, len(u.Builds))
	for name := range u.Builds {
		ret = append(ret, name)
	}
	sort.Strings(ret)
	return ret
}

func (u *Unit) Template(ctx context.Context, name, squadron, unit, namespace string, global map[string]any, helmArgs []string) ([]byte, error) {
	var ret bytes.Buffer
	valueBytes, err := u.ValuesYAML(global)
	if err != nil {
		return nil, err
	}

	cmd := util.NewHelmCommand().Args("template", name).
		Stdin(bytes.NewReader(valueBytes)).
		Stdout(&ret).
		Args("--dependency-update").
		Args("--namespace", namespace).
		Args("--debug").
		Args("--set", "global.foomo.squadron.name="+squadron).
		Args("--set", "global.foomo.squadron.unit="+unit).
		Args(u.PostRendererArgs()...).
		Args("--values", "-").
		Args(helmArgs...)

	if strings.HasPrefix(u.Chart.Repository, "file://") {
		cmd.Args(path.Clean(strings.TrimPrefix(u.Chart.Repository, "file://")))
	} else {
		cmd.Args(u.Chart.Name)
		if u.Chart.Repository != "" {
			cmd.Args("--repo", u.Chart.Repository)
		}
		if u.Chart.Version != "" {
			cmd.Args("--version", u.Chart.Version)
		}
	}
	if out, err := cmd.Run(ctx); err != nil {
		return nil, errors.Wrap(err, out)
	}

	return ret.Bytes(), nil
}

func (u *Unit) PostRendererArgs() []string {
	var ret []string
	if u.Kustomize != "" {
		ret = append(ret,
			"--post-renderer", "squadron",
			"--post-renderer-args", "post-renderer",
			"--post-renderer-args", u.Kustomize,
		)
	}
	return ret
}
