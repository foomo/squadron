package config

import (
	"context"
	"fmt"

	"github.com/foomo/squadron/internal/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Build struct {
	Args         []string `yaml:"args,omitempty"`
	Builder      string   `yaml:"builder,omitempty"`
	CacheFrom    string   `yaml:"cache_from,omitempty"`
	CacheTo      string   `yaml:"cache_to,omitempty"`
	Context      string   `yaml:"context,omitempty"`
	Dockerfile   string   `yaml:"dockerfile,omitempty"`
	ExtraHosts   []string `yaml:"extra_hosts,omitempty"`
	Image        string   `yaml:"image,omitempty"`
	IIDFile      string   `yaml:"iidfile,omitempty"`
	Labels       []string `yaml:"labels,omitempty"`
	Load         bool     `yaml:"load,omitempty"`
	MetadataFile string   `yaml:"metadata_file,omitempty"`
	Network      string   `yaml:"network,omitempty"`
	NoCache      bool     `yaml:"no_cache,omitempty"`
	Output       string   `yaml:"output,omitempty"`
	Platform     string   `yaml:"platform,omitempty"`
	Platforms    []string `yaml:"platforms,omitempty"`
	Secrets      []string `yaml:"secrets,omitempty"`
	ShmSize      string   `yaml:"shm_size,omitempty"`
	SSH          string   `yaml:"ssh,omitempty"`
	Tag          string   `yaml:"tag,omitempty"`
	Target       string   `yaml:"target,omitempty"`
	ULimit       string   `yaml:"ulimit,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (b *Build) Build(ctx context.Context, args []string) (string, error) {
	logrus.Debugf("running docker build for %q", b.Context)
	return util.NewDockerCommand().Build(b.Context).
		ListArg("--add-host", b.ExtraHosts).
		ListArg("--build-arg", b.Args).
		Arg("--builder", b.Builder).
		Arg("--cache-from", b.CacheFrom).
		Arg("--cache-to", b.CacheTo).
		Arg("--file", b.Dockerfile).
		Arg("--iidfile", b.IIDFile).
		ListArg("--label", b.Labels).
		BoolArg("--load", b.Load).
		Arg("--metadata-file", b.MetadataFile).
		Arg("--network", b.Network).
		BoolArg("--no-cache", b.NoCache).
		Arg("--output", b.Output).
		Arg("--platform", b.Platform).
		ListArg("--platform", b.Platforms).
		// Arg("--progress", xxx).
		// Arg("--push", xxx).
		// Arg("--pull", xxx).
		// Arg("--quiet", xxx).
		ListArg("--secret", b.Secrets).
		Arg("--shm-size", b.ShmSize).
		Arg("--ssh", b.SSH).
		Arg("--tag", fmt.Sprintf("%s:%s", b.Image, b.Tag)).
		Arg("--target", b.Target).
		Arg("--ulimit", b.ULimit).
		Args(args...).
		Run(ctx)
}

func (b *Build) Push(ctx context.Context, args []string) (string, error) {
	logrus.Debugf("running docker push for %s:%s", b.Image, b.Tag)
	return util.NewDockerCommand().Push(b.Image, b.Tag).Args(args...).Run(ctx)
}

func (b *Build) UnmarshalYAML(value *yaml.Node) error {
	switch value.Tag {
	case "!!map":
		type wrapper Build
		return value.Decode((*wrapper)(b))
	case "!!str":
		var vString string
		if err := value.Decode(&vString); err != nil {
			return err
		}
		b.Context = vString
		return nil
	default:
		return fmt.Errorf("unsupported node tag type for %T: %q", b, value.Tag)
	}
}
