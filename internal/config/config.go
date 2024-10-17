package config

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Version of the schema
	Version string `json:"version,omitempty" yaml:"version,omitempty" jsonschema:"pattern=^[0-9]\\.[0-9]$,required"`
	// Global values to be injected into all squadron values
	Vars map[string]any `json:"vars,omitempty" yaml:"vars,omitempty"`
	// Global values to be injected into all squadron values
	Global map[string]any `json:"global,omitempty" yaml:"global,omitempty"`
	// Global builds that can be referenced as dependencies
	Builds map[string]Build `json:"builds,omitempty" yaml:"builds,omitempty"`
	// Squadron definitions
	Squadrons Map[Map[*Unit]] `json:"squadron,omitempty" yaml:"squadron,omitempty"`
}

// JSONSchemaProperty type workaround
func (Config) JSONSchemaProperty(prop string) any {
	if prop == "squadron" {
		return map[string]map[string]*Unit{}
	}
	return nil
}

// BuildDependencies returns a map of requested build dependencies
func (c *Config) BuildDependencies(ctx context.Context) map[string]Build {
	ret := map[string]Build{}
	_ = c.Squadrons.Iterate(ctx, func(ctx context.Context, key string, value Map[*Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *Unit) error {
			for _, build := range v.Builds {
				for _, dependency := range build.Dependencies {
					b, ok := c.Builds[dependency]
					if !ok {
						return errors.Errorf("missing build dependency `%s`", dependency)
					}
					ret[dependency] = b
				}
			}
			return nil
		})
	})
	if len(ret) > 0 {
		return ret
	}
	return nil
}

// Trim delete empty squadron recursively
func (c *Config) Trim(ctx context.Context) {
	_ = c.Squadrons.Iterate(ctx, func(ctx context.Context, key string, value Map[*Unit]) error {
		value.Trim()
		return nil
	})
	c.Squadrons.Trim()
}

// UnmarshalYAML interface method
func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!map":
		type wrapper Config
		return value.Decode((*wrapper)(c))
	default:
		return fmt.Errorf("unsupported node tag type for %T: %q", c, value.Tag)
	}
}
