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
	// Build context
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
	// AddHost add a custom host-to-IP mapping (format: "host:ip")
	AddHost []string `json:"add_host,omitempty" yaml:"add_host,omitempty"`
	// Allow extra privileged entitlement (e.g., "network.host", "security.insecure")
	Allow []string `json:"allow,omitempty" yaml:"allow,omitempty"`
	// Attest parameters (format: "type=sbom,generator=image")
	Attest []string `json:"attest,omitempty" yaml:"attest,omitempty"`
	// BuildArg set build-time variables
	BuildArg []string `json:"build_arg,omitempty" yaml:"build_arg,omitempty"`
	// BuildContext additional build contexts (e.g., name=path)
	BuildContext []string `json:"build_context,omitempty" yaml:"build_context,omitempty"`
	// Builder override the configured builder instance
	Builder string `json:"builder,omitempty" yaml:"builder,omitempty"`
	// CacheFrom external cache sources (e.g., "user/app:cache", "type=local,src=path/to/dir")
	CacheFrom string `json:"cache_from,omitempty" yaml:"cache_from,omitempty"`
	// CacheTo cache export destinations (e.g., "user/app:cache", "type=local,dest=path/to/dir")
	CacheTo string `json:"cache_to,omitempty" yaml:"cache_to,omitempty"`
	// CGroupParent optional parent cgroup for the container
	CGroupParent string `json:"cgroup_parent,omitempty" yaml:"cgroup_parent,omitempty"`
	// File name of the Dockerfile (default: "PATH/Dockerfile")
	File string `json:"file,omitempty" yaml:"file,omitempty"`
	// IIDFile write the image ID to the file
	IIDFile string `json:"iidfile,omitempty" yaml:"iidfile,omitempty"`
	// Label wet metadata for an image
	Label []string `json:"label,omitempty" yaml:"label,omitempty"`
	// Load shorthand for "--output=type=docker"
	Load bool `json:"load,omitempty" yaml:"load,omitempty"`
	// MetadataFile write build result metadata to the file
	MetadataFile string `json:"metadata_file,omitempty" yaml:"metadata_file,omitempty"`
	// Network set the networking mode for the "RUN" instructions during build (default "default")
	Network string `json:"network,omitempty" yaml:"network,omitempty"`
	// NoCache do not use cache when building the image
	NoCache bool `json:"no_cache,omitempty" yaml:"no_cache,omitempty"`
	// NoCacheFilter do not cache specified stages
	NoCacheFilter []string `json:"no_cache_filter,omitempty" yaml:"no_cache_filter,omitempty"`
	// Output destination (format: "type=local,dest=path")
	Output string `json:"output,omitempty" yaml:"output,omitempty"`
	// Platform set target platform for build
	Platform string `json:"platform,omitempty" yaml:"platform,omitempty"`
	// Secret to expose to the build (format: "id=mysecret[,src=/local/secret]")
	Secret []string `json:"secret,omitempty" yaml:"secret,omitempty"`
	// ShmSize size of "/dev/shm"
	ShmSize string `json:"shm_size,omitempty" yaml:"shm_size,omitempty"`
	// SSH agent socket or keys to expose to the build (format: "default|<id>[=<socket>|<key>[,<key>]]")
	SSH string `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	// Tag name and optionally a tag (format: "name:tag")
	Tag string `json:"tag,omitempty" yaml:"tag,omitempty"`
	// Image name
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	// Target set the target build stage to build
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
	// ULimit ulimit options (default [])
	ULimit string `json:"ulimit,omitempty" yaml:"ulimit,omitempty"`
	// Dependencies list of build names defined in the squadron configuration
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	// Suppress the build output and print image ID on succes
	Quiet bool `json:"quiet,omitempty" yaml:"quiet,omitempty"`
	// Always attempt to pull all referenced images
	Pull bool `json:"pull,omitempty" yaml:"pull,omitempty"`
	// Shorthand for "--output=type=registry"
	Push bool `json:"push,omitempty" yaml:"push,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (b *Build) Build(ctx context.Context, squadron, unit string, args []string) (string, error) {
	var cleanArgs []string
	for _, arg := range args {
		if value, err := util.RenderTemplateString(arg, map[string]any{"Squadron": squadron, "Unit": unit, "Build": b}); err != nil {
			return "", err
		} else {
			cleanArgs = append(cleanArgs, strings.Split(value, " ")...)
		}
	}

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

	pterm.Debug.Printfln("running docker build for %q", b.Context)

	return util.NewDockerCommand().Build(b.Context).
		TemplateData(map[string]string{"image": b.Image, "tag": b.Tag}).
		ListArg("--add-host", b.AddHost).
		ListArg("--add-host", b.AddHost).
		ListArg("--allow", b.Allow).
		ListArg("--attest", b.Attest).
		ListArg("--build-arg", b.BuildArg).
		ListArg("--build-context", b.BuildContext).
		Arg(argOverride("--builder", b.Builder, args)).
		Arg(argOverride("--cache-from", b.CacheFrom, args)).
		Arg(argOverride("--cache-to", b.CacheTo, args)).
		Arg(argOverride("--file", b.File, args)).
		Arg(argOverride("--iidfile", b.IIDFile, args)).
		ListArg("--label", b.Label).
		BoolArg(boolArgOverride("--load", b.Load, args)).
		Arg(argOverride("--metadata-file", b.MetadataFile, args)).
		Arg(argOverride("--network", b.Network, args)).
		BoolArg(boolArgOverride("--no-cache", b.NoCache, args)).
		ListArg("--noe-cache-filter", b.NoCacheFilter).
		Arg(argOverride("--output", b.Output, args)).
		Arg(argOverride("--platform", b.Platform, args)).
		// Arg("--progress", xxx).
		// Arg("--provenance", xxx).
		BoolArg(boolArgOverride("--push", b.Push, args)).
		BoolArg(boolArgOverride("--pull", b.Pull, args)).
		BoolArg(boolArgOverride("--quiet", b.Quiet, args)).
		ListArg("--secret", b.Secret).
		Arg(argOverride("--shm-size", b.ShmSize, args)).
		Arg(argOverride("--ssh", b.SSH, args)).
		Arg(argOverride("--tag", fmt.Sprintf("%s:%s", b.Image, b.Tag), args)).
		Arg(argOverride("--target", b.Target, args)).
		Arg(argOverride("--ulimit", b.ULimit, args)).
		Args(cleanArgs...).Run(ctx)
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
