package squadron

import (
	"context"
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
func (u *Unit) Build(ctx context.Context) error {
	for _, build := range u.Builds {
		if err := build.Build(ctx); err != nil {
			return err
		}
	}
	return nil
}

// Push ...
func (u *Unit) Push(ctx context.Context) error {
	for _, build := range u.Builds {
		if err := build.Push(ctx); err != nil {
			return err
		}
	}
	return nil
}
