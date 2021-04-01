package squadron

import (
	"fmt"

	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
	"github.com/sirupsen/logrus"
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
	logrus.Infof("checking image exists for %s:%s", b.Image, b.Tag)
	return util.NewDockerCommand().ImageExists(b.Image, b.Tag)
}

// Build ...
func (b *Build) Build() error {
	logrus.Infof("running docker build for %q", b.Context)
	cmd := util.NewDockerCommand()
	cmd.Args("-t", fmt.Sprintf("%s:%s", b.Image, b.Tag)).
		Args("--file", b.Dockerfile).
		ListArg("--build-arg", b.Args).
		ListArg("--label", b.Labels).
		ListArg("--cache-from", b.CacheFrom).
		Args("--network", b.Network).
		Args("--target", b.Target).
		Args("--shm-size", b.ShmSize).
		ListArg("--add-host", b.ExtraHosts).
		Args("--isolation", b.Isolation)
	_, err := cmd.Build(b.Context)
	return err
}

// Push ...
func (b *Build) Push() error {
	logrus.Infof("running docker push for %s:%s", b.Image, b.Tag)
	_, err := util.NewDockerCommand().Push(b.Image, b.Tag)
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
