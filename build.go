package squadron

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Build struct {
	Image      string   `yaml:"image,omitempty"`
	Tag        string   `yaml:"tag,omitempty"`
	Context    string   `yaml:"context,omitempty"`
	Dockerfile string   `yaml:"dockerfile,omitempty"`
	Args       []string `yaml:"args,omitempty"`
	Labels     []string `yaml:"labels,omitempty"`
	CacheFrom  []string `yaml:"cache_from,omitempty"`
	Network    string   `yaml:"network,omitempty"`
	Target     string   `yaml:"target,omitempty"`
	ShmSize    string   `yaml:"shm_size,omitempty"`
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`
	Isolation  string   `yaml:"isolation,omitempty"`
}

func (b *Build) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == "!!map" {
		type wrapper Build
		return value.Decode((*wrapper)(b))
	}
	if value.Tag == "!!str" {
		var vString string
		if err := value.Decode(&vString); err != nil {
			return err
		}
		b.Context = vString
	}
	return fmt.Errorf("unsupported node tag type for %T: %q", b, value.Tag)
}
