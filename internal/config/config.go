package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version   string                 `yaml:"version,omitempty"`
	Global    map[string]interface{} `yaml:"global,omitempty"`
	Squadrons Map[Map[*Unit]]        `yaml:"squadron,omitempty"`
}

func (c *Config) Trim() {
	_ = c.Squadrons.Iterate(func(key string, value Map[*Unit]) error {
		value.Trim()
		return nil
	})
	c.Squadrons.Trim()
}

func (c *Config) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!map":
		type wrapper Config
		return value.Decode((*wrapper)(c))
	default:
		return fmt.Errorf("unsupported node tag type for %T: %q", c, value.Tag)
	}
}
