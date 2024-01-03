package config

import (
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"

	"github.com/foomo/squadron/internal/helm"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	yamlv2 "gopkg.in/yaml.v2"
)

type Unit struct {
	Chart  helm.Dependency        `yaml:"chart,omitempty"`
	Tags   []Tag                  `yaml:"tags,omitempty"`
	Builds map[string]Build       `yaml:"builds,omitempty"`
	Values map[string]interface{} `yaml:"values,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (u *Unit) ValuesYAML(global map[string]interface{}) ([]byte, error) {
	values := u.Values
	if values == nil {
		values = map[string]interface{}{}
	}
	if global != nil {
		if _, ok := values["global"]; !ok {
			values["global"] = global
		}
	}
	return yamlv2.Marshal(values)
}

func (u *Unit) Build(ctx context.Context, squadron, unit string, args []string) (string, error) {
	var i int
	for _, build := range u.Builds {
		i++
		pterm.Info.Printfln("[%d/%d] Building %s/%s", i, len(u.Builds), squadron, unit)
		pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
		if out, err := build.Build(ctx, args); err != nil {
			pterm.Error.Printfln("[%d/%d] Failed to build squadron unit %s/%s", i, len(u.Builds), squadron, unit)
			pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
			return out, err
		}
	}
	return "", nil
}

func (u *Unit) Push(ctx context.Context, squadron, unit string, args []string) (string, error) {
	var i int
	for _, build := range u.Builds {
		i++
		pterm.Info.Printfln("[%d/%d] Pushing %s/%s", i, len(u.Builds), squadron, unit)
		pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
		if out, err := build.Push(ctx, args); err != nil {
			pterm.Error.Printfln("[%d/%d] Failed to push %s/%s", i, len(u.Builds), squadron, unit)
			pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
			return out, err
		}
	}
	return "", nil
}

func (u *Unit) Template(ctx context.Context, squadron, unit, namespace string, global map[string]interface{}, helmArgs []string) ([]byte, error) {
	var ret bytes.Buffer
	valueBytes, err := u.ValuesYAML(global)
	if err != nil {
		return nil, err
	}

	cmd := util.NewHelmCommand().Args("template", unit).
		Stdin(bytes.NewReader(valueBytes)).
		Stdout(&ret).
		Args("--dependency-update").
		Args("--namespace", namespace).
		Args("--debug").
		Args("--set", fmt.Sprintf("squadron=%s", squadron)).
		Args("--set", fmt.Sprintf("unit=%s", unit)).
		Args("--values", "-").
		Args(helmArgs...)
	if strings.HasPrefix(u.Chart.Repository, "file://") {
		cmd.Args(path.Clean(strings.TrimPrefix(u.Chart.Repository, "file://")))
	} else {
		cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository, "--version", u.Chart.Version)
	}
	if out, err := cmd.Run(ctx); err != nil {
		return nil, errors.Wrap(err, out)
	}

	return ret.Bytes(), nil
}

func (u *Unit) DependencyUpdate(ctx context.Context) error {
	// update local chart dependencies
	// https://stackoverflow.com/questions/59210148/error-found-in-chart-yaml-but-missing-in-charts-directory-mysql
	if strings.HasPrefix(u.Chart.Repository, "file:///") {
		pterm.Debug.Printfln("running helm dependency update for %s", u.Chart.Repository)
		sh := util.NewHelmCommand().
			Cwd(strings.TrimPrefix(u.Chart.Repository, "file://")).
			Args("dependency", "update", "--skip-refresh", "--debug")
		if out, err := sh.Run(ctx); err != nil {
			return errors.Wrap(err, out)
		} else {
			pterm.Debug.Println(out)
		}
	}
	return nil
}
