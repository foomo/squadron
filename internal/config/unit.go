package config

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/foomo/squadron/internal/helm"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	yamlv2 "gopkg.in/yaml.v2"
)

type Unit struct {
	Chart     helm.Dependency  `json:"chart,omitempty" yaml:"chart,omitempty"`
	Kustomize string           `json:"kustomize,omitempty" yaml:"kustomize,omitempty"`
	Tags      Tags             `json:"tags,omitempty" yaml:"tags,omitempty"`
	Builds    map[string]Build `json:"builds,omitempty" yaml:"builds,omitempty"`
	Values    map[string]any   `json:"values,omitempty" yaml:"values,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (u *Unit) ValuesYAML(global, vars map[string]any) ([]byte, error) {
	values := u.Values
	if values == nil {
		values = map[string]any{}
	}
	if global != nil {
		if _, ok := values["global"]; !ok {
			values["global"] = global
		}
	}
	if vars != nil {
		if _, ok := values["vars"]; !ok {
			values["vars"] = vars
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

func (u *Unit) Template(ctx context.Context, name, squadron, unit, namespace string, global, vars map[string]any, helmArgs []string) ([]byte, error) {
	var ret bytes.Buffer
	valueBytes, err := u.ValuesYAML(global, vars)
	if err != nil {
		return nil, err
	}

	cmd := util.NewHelmCommand().Args("template", name).
		Stdin(bytes.NewReader(valueBytes)).
		Stdout(&ret).
		Args("--dependency-update").
		Args("--namespace", namespace).
		Args("--debug").
		Args("--set", fmt.Sprintf("squadron=%s", squadron)).
		Args("--set", fmt.Sprintf("unit=%s", unit)).
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
