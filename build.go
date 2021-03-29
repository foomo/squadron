package squadron

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
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

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (b *Build) Exists() (bool, error) {
	return util.NewDockerCommand(logrus.NewEntry(logrus.StandardLogger())).ImageExists(b.Image, b.Tag)
}

// Build ...
func (b *Build) Build() error {
	cmd := util.NewDockerCommand(logrus.NewEntry(logrus.StandardLogger()))
	cmd.Option("-t", fmt.Sprintf("%s:%s", b.Image, b.Tag))
	cmd.Option("--file", b.Dockerfile)
	cmd.ListOption("--build-arg", b.Args)
	cmd.ListOption("--label", b.Labels)
	cmd.ListOption("--cache-from", b.CacheFrom)
	cmd.Option("--network", b.Network)
	cmd.Option("--target", b.Target)
	cmd.Option("--shm-size", b.ShmSize)
	cmd.ListOption("--add-host", b.ExtraHosts)
	cmd.Option("--isolation", b.Isolation)
	_, err := cmd.Build(b.Context)
	return err
}

// Push ...
func (b *Build) Push() error {
	cmd := util.NewDockerCommand(logrus.NewEntry(logrus.StandardLogger()))
	_, err := cmd.Push(b.Image, b.Tag)
	return err
}

// UnmarshalYAML ...
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
