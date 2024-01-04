package config

import (
	"fmt"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Version of the schema
	Version string `yaml:"version,omitempty"`
	// Global values to be injected into all squadron values
	Global map[string]interface{} `yaml:"global,omitempty"`
	// Global builds that can be referenced as dependencies
	Builds map[string]Build `yaml:"builds,omitempty"`
	// Squadron definitions
	Squadrons Map[Map[*Unit]] `yaml:"squadron,omitempty"`
}

// BuildDependencies returns a map of requested build dependencies
func (c *Config) BuildDependencies() map[string]Build {
	ret := map[string]Build{}
	_ = c.Squadrons.Iterate(func(key string, value Map[*Unit]) error {
		return value.Iterate(func(k string, v *Unit) error {
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
func (c *Config) Trim() {
	_ = c.Squadrons.Iterate(func(key string, value Map[*Unit]) error {
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
