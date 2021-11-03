package squadron

import (
	"context"

	"github.com/pterm/pterm"
)

type Unit struct {
	Chart  ChartDependency        `yaml:"chart,omitempty"`
	Builds map[string]Build       `yaml:"builds,omitempty"`
	Values map[string]interface{} `yaml:"values,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

// Build ...
func (u *Unit) Build(ctx context.Context, squadron, unit string) (string, error) {
	var i int
	for _, build := range u.Builds {
		i++
		pterm.Info.Printfln("[%d/%d] Building %s/%s", i, len(u.Builds), squadron, unit)
		pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
		if out, err := build.Build(ctx); err != nil {
			pterm.Error.Printfln("[%d/%d] Failed to build squadron unit %s/%s", i, len(u.Builds), squadron, unit)
			pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
			return out, err
		}
	}
	return "", nil
}

// Push ...
func (u *Unit) Push(ctx context.Context, squadron, unit string) (string, error) {
	var i int
	for _, build := range u.Builds {
		i++
		pterm.Info.Printfln("[%d/%d] Pushing %s/%s", i, len(u.Builds), squadron, unit)
		pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
		if out, err := build.Push(ctx); err != nil {
			pterm.Error.Printfln("[%d/%d] Failed to push %s/%s", i, len(u.Builds), squadron, unit)
			pterm.FgGray.Printfln("└ %s:%s", build.Image, build.Tag)
			return out, err
		}
	}
	return "", nil
}
