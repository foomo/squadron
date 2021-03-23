package squadron

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

type Build struct {
	Image      string   `yaml:"image"`
	Tag        string   `yaml:"tag"`
	Context    string   `yaml:"context"`
	Dockerfile string   `yaml:"dockerfile"`
	Args       []string `yaml:"args"`
	Labels     []string `yaml:"labels"`
	CacheFrom  []string `yaml:"cache_from"`
	Network    string   `yaml:"network"`
	Target     string   `yaml:"target"`
	ShmSize    string   `yaml:"shm_size"`
	ExtraHosts []string `yaml:"extra_hosts"`
	Isolation  string   `yaml:"isolation"`
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
