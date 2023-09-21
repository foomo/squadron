package config

import (
	"context"
	"os"
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
		values["global"] = global
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

func (u *Unit) DependencyUpdate(ctx context.Context) error {
	// update local chart dependencies
	// https://stackoverflow.com/questions/59210148/error-found-in-chart-yaml-but-missing-in-charts-directory-mysql
	if strings.HasPrefix(u.Chart.Repository, "file:///") {
		pterm.Debug.Printfln("running helm dependency update for %s", u.Chart.Repository)
		if out, err := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("dependency", "update").
			Cwd(strings.TrimPrefix(u.Chart.Repository, "file://")).
			Run(ctx); err != nil {
			return errors.Wrap(err, out)
		}
	}
	return nil
}
