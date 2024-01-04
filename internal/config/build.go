package config

import (
	"context"
	"fmt"

	"github.com/foomo/squadron/internal/util"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

type Build struct {
	Context string `yaml:"context,omitempty"`
	// AddHost add a custom host-to-IP mapping (format: "host:ip")
	AddHost []string `yaml:"add_host,omitempty"`
	// Allow extra privileged entitlement (e.g., "network.host", "security.insecure")
	Allow []string `yaml:"allow,omitempty"`
	// Attest parameters (format: "type=sbom,generator=image")
	Attest []string `yaml:"attest,omitempty"`
	// BuildArg set build-time variables
	BuildArg []string `yaml:"build_arg,omitempty"`
	// BuildContext additional build contexts (e.g., name=path)
	BuildContext []string `yaml:"build_context,omitempty"`
	// Builder override the configured builder instance
	Builder string `yaml:"builder,omitempty"`
	// CacheFrom external cache sources (e.g., "user/app:cache", "type=local,src=path/to/dir")
	CacheFrom string `yaml:"cache_from,omitempty"`
	// CacheTo cache export destinations (e.g., "user/app:cache", "type=local,dest=path/to/dir")
	CacheTo string `yaml:"cache_to,omitempty"`
	// CGroupParent optional parent cgroup for the container
	CGroupParent string `yaml:"cgroup_parent,omitempty"`
	// File name of the Dockerfile (default: "PATH/Dockerfile")
	File string `yaml:"file,omitempty"`
	// IIDFile write the image ID to the file
	IIDFile string `yaml:"iidfile,omitempty"`
	// Label wet metadata for an image
	Label []string `yaml:"label,omitempty"`
	// Load shorthand for "--output=type=docker"
	Load bool `yaml:"load,omitempty"`
	// MetadataFile write build result metadata to the file
	MetadataFile string `yaml:"metadata_file,omitempty"`
	// Network set the networking mode for the "RUN" instructions during build (default "default")
	Network string `yaml:"network,omitempty"`
	// NoCache do not use cache when building the image
	NoCache bool `yaml:"no_cache,omitempty"`
	// NoCacheFilter do not cache specified stages
	NoCacheFilter []string `yaml:"no_cache_filter,omitempty"`
	// Output destination (format: "type=local,dest=path")
	Output string `yaml:"output,omitempty"`
	// Platform set target platform for build
	Platform string `yaml:"platform,omitempty"`
	// Secret to expose to the build (format: "id=mysecret[,src=/local/secret]")
	Secret []string `yaml:"secret,omitempty"`
	// ShmSize size of "/dev/shm"
	ShmSize string `yaml:"shm_size,omitempty"`
	// SSH agent socket or keys to expose to the build (format: "default|<id>[=<socket>|<key>[,<key>]]")
	SSH string `yaml:"ssh,omitempty"`
	// Tag name and optionally a tag (format: "name:tag")
	Tag   string `yaml:"tag,omitempty"`
	Image string `yaml:"image,omitempty"`
	// Target set the target build stage to build
	Target string `yaml:"target,omitempty"`
	// ULimit ulimit options (default []
	ULimit string `yaml:"ulimit,omitempty"`
	// Dependencies list of build names defined in the squadron configuration
	Dependencies []string `yaml:"dependencies,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (b *Build) Build(ctx context.Context, args []string) (string, error) {
	pterm.Debug.Printfln("running docker build for %q", b.Context)
	return util.NewDockerCommand().Build(b.Context).
		ListArg("--add-host", b.AddHost).
		ListArg("--allow", b.Allow).
		ListArg("--attest", b.Attest).
		ListArg("--build-arg", b.BuildArg).
		ListArg("--build-contet", b.BuildContext).
		Arg("--builder", b.Builder).
		Arg("--cache-from", b.CacheFrom).
		Arg("--cache-to", b.CacheTo).
		Arg("--file", b.File).
		Arg("--iidfile", b.IIDFile).
		ListArg("--label", b.Label).
		BoolArg("--load", b.Load).
		Arg("--metadata-file", b.MetadataFile).
		Arg("--network", b.Network).
		BoolArg("--no-cache", b.NoCache).
		ListArg("--noe-cache-filter", b.NoCacheFilter).
		Arg("--output", b.Output).
		Arg("--platform", b.Platform).
		// Arg("--progress", xxx).
		// Arg("--provenance", xxx).
		// Arg("--push", xxx).
		// Arg("--pull", xxx).
		// Arg("--quiet", xxx).
		ListArg("--secret", b.Secret).
		Arg("--shm-size", b.ShmSize).
		Arg("--ssh", b.SSH).
		Arg("--tag", fmt.Sprintf("%s:%s", b.Image, b.Tag)).
		Arg("--target", b.Target).
		Arg("--ulimit", b.ULimit).
		Args(args...).
		Run(ctx)
}

func (b *Build) Push(ctx context.Context, args []string) (string, error) {
	pterm.Debug.Printfln("running docker push for %s:%s", b.Image, b.Tag)
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
