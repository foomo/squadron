package squadron

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Name    string                 `yaml:"name,omitempty"`
	Version string                 `yaml:"version,omitempty"`
	Prefix  string                 `yaml:"prefix,omitempty"`
	Unite   bool                   `yaml:"unite,omitempty"`
	Global  map[string]interface{} `yaml:"global,omitempty"`
	Units   Units                  `yaml:"squadron,omitempty"`
}

// UnmarshalYAML ...
func (c *Configuration) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == TagMap {
		type wrapper Configuration
		err := value.Decode((*wrapper)(c))
		if err == nil {
			// if the decode is successful, remove units that are nil
			c.removeNilUnits()
		}
		return err
	}
	return fmt.Errorf("unsupported node tag type for %T: %q", c, value.Tag)
}

func (c *Configuration) removeNilUnits() {
	for uName, u := range c.Units {
		if u == nil {
			delete(c.Units, uName)
		}
	}
}
