package config

import (
	"context"
	"fmt"
	"slices"
	"strings"

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
	argOverride := func(name string, vs string, args []string) (string, string) {
		if slices.ContainsFunc(args, func(s string) bool {
			return strings.HasPrefix(s, name)
		}) {
			return "", ""
		}
		return name, vs
	}
	boolArgOverride := func(name string, vs bool, args []string) (string, bool) {
		if slices.ContainsFunc(args, func(s string) bool {
			return strings.HasPrefix(s, name)
		}) {
			return "", false
		}
		return name, vs
	}
	listArgOverride := func(name string, vs, args []string) (string, []string) {
		if slices.ContainsFunc(args, func(s string) bool {
			return strings.HasPrefix(s, name)
		}) {
			return "", nil
		}
		return name, vs
	}

	pterm.Debug.Printfln("running docker build for %q", b.Context)
	return util.NewDockerCommand().Build(b.Context).
		ListArg(listArgOverride("--add-host", b.AddHost, args)).
		ListArg(listArgOverride("--allow", b.Allow, args)).
		ListArg(listArgOverride("--attest", b.Attest, args)).
		ListArg(listArgOverride("--build-arg", b.BuildArg, args)).
		ListArg(listArgOverride("--build-contet", b.BuildContext, args)).
		Arg(argOverride("--builder", b.Builder, args)).
		Arg(argOverride("--cache-from", b.CacheFrom, args)).
		Arg(argOverride("--cache-to", b.CacheTo, args)).
		Arg(argOverride("--file", b.File, args)).
		Arg(argOverride("--iidfile", b.IIDFile, args)).
		ListArg(listArgOverride("--label", b.Label, args)).
		BoolArg(boolArgOverride("--load", b.Load, args)).
		Arg(argOverride("--metadata-file", b.MetadataFile, args)).
		Arg(argOverride("--network", b.Network, args)).
		BoolArg(boolArgOverride("--no-cache", b.NoCache, args)).
		ListArg(listArgOverride("--noe-cache-filter", b.NoCacheFilter, args)).
		Arg(argOverride("--output", b.Output, args)).
		Arg(argOverride("--platform", b.Platform, args)).
		// Arg("--progress", xxx).
		// Arg("--provenance", xxx).
		// Arg("--push", xxx).
		// Arg("--pull", xxx).
		// Arg("--quiet", xxx).
		ListArg(listArgOverride("--secret", b.Secret, args)).
		Arg(argOverride("--shm-size", b.ShmSize, args)).
		Arg(argOverride("--ssh", b.SSH, args)).
		Arg(argOverride("--tag", fmt.Sprintf("%s:%s", b.Image, b.Tag), args)).
		Arg(argOverride("--target", b.Target, args)).
		Arg(argOverride("--ulimit", b.ULimit, args)).
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
