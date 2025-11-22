package config

import (
	"context"
	"strings"

	"github.com/foomo/squadron/internal/qflag"
	"github.com/foomo/squadron/internal/util"
	"github.com/pterm/pterm"
	"github.com/spf13/pflag"
)

type Build struct {
	// Build context
	Context string `json:"context,omitempty" yaml:"context,omitempty"`
	// Dependencies list of build names defined in the squadron configuration
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`

	// AddHost add a custom host-to-IP mapping (format: "host:ip")
	AddHost []string `json:"add_host,omitempty" yaml:"add_host,omitempty"`
	// Allow extra privileged entitlement (e.g., "network.host", "security.insecure")
	Allow []string `json:"allow,omitempty" yaml:"allow,omitempty"`
	// Add annotation to the image
	Annotation []string `json:"annotation,omitempty" yaml:"annotation,omitempty"`
	// Attest parameters (format: "type=sbom,generator=image")
	Attest []string `json:"attest,omitempty" yaml:"attest,omitempty"`
	// BuildArg set build-time variables
	BuildArg []string `json:"build_arg,omitempty" yaml:"build_arg,omitempty"`
	// BuildContext additional build contexts (e.g., name=path)
	BuildContext []string `json:"build_context,omitempty" yaml:"build_context,omitempty"`
	// Builder override the configured builder instance
	Builder string `json:"builder,omitempty" yaml:"builder,omitempty"`
	// CacheFrom external cache sources (e.g., "user/app:cache", "type=local,src=path/to/dir")
	CacheFrom []string `json:"cache_from,omitempty" yaml:"cache_from,omitempty"`
	// CacheTo cache export destinations (e.g., "user/app:cache", "type=local,dest=path/to/dir")
	CacheTo []string `json:"cache_to,omitempty" yaml:"cache_to,omitempty"`
	// Set method for evaluating build ("check", "outline", "targets") (default "build")
	Call string `json:"call,omitempty" yaml:"call,omitempty"`
	// CGroupParent optional parent cgroup for the container
	CGroupParent string `json:"cgroup_parent,omitempty" yaml:"cgroup_parent,omitempty"`
	// Shorthand for "--call=check"
	Check bool `json:"check,omitempty" yaml:"check,omitempty"`
	// Enable debug logging
	Debug bool `json:"debug,omitempty" yaml:"debug,omitempty"`
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
	Output []string `json:"output,omitempty" yaml:"output,omitempty"`
	// Platform set target platform for build
	Platform []string `json:"platform,omitempty" yaml:"platform,omitempty"`
	// Set type of progress output ("auto", "none",  "plain", "quiet", "rawjson", "tty"). Use plain to show container output (default "auto")
	Progress string `json:"progress,omitempty" yaml:"progress,omitempty"`
	// Shorthand for "--attest=type=provenance"
	Provenance bool `json:"provenance,omitempty" yaml:"provenance,omitempty"`
	// Always attempt to pull all referenced images
	Pull bool `json:"pull,omitempty" yaml:"pull,omitempty"`
	// Shorthand for "--output=type=registry"
	Push bool `json:"push,omitempty" yaml:"push,omitempty"`
	// Suppress the build output and print image ID on succes
	Quiet bool `json:"quiet,omitempty" yaml:"quiet,omitempty"`
	// Shorthand for "--attest=type=sbom"
	Sbom bool `json:"sbom,omitempty" yaml:"sbom,omitempty"`
	// Secret to expose to the build (format: "id=mysecret[,src=/local/secret]")
	Secret []string `json:"secret,omitempty" yaml:"secret,omitempty"`
	// ShmSize size of "/dev/shm"
	ShmSize string `json:"shm_size,omitempty" yaml:"shm_size,omitempty"`
	// SSH agent socket or keys to expose to the build (format: "default|<id>[=<socket>|<key>[,<key>]]")
	SSH []string `json:"ssh,omitempty" yaml:"ssh,omitempty"`
	// Tag name and optionally a tag (format: "name:tag")
	Tag []string `json:"tag,omitempty" yaml:"tag,omitempty"`
	// Target set the target build stage to build
	Target string `json:"target,omitempty" yaml:"target,omitempty"`
	// ULimit ulimit options (default [])
	ULimit string `json:"ulimit,omitempty" yaml:"ulimit,omitempty"`
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (b *Build) Build(ctx context.Context, squadron, unit string, args []string) (string, error) {
	var cleanArgs []string

	for _, arg := range args {
		cleanArgs = append(cleanArgs, strings.Split(arg, " ")...)
	}

	f := pflag.NewFlagSet("build", pflag.ContinueOnError)
	f.StringSlice("add-host", b.AddHost, "")
	f.StringArray("allow", b.Allow, "")
	f.StringArray("annotation", b.Annotation, "")
	f.StringArray("attest", b.Attest, "")
	f.StringArray("build-arg", b.BuildArg, "")
	f.StringArray("build-context", b.BuildContext, "")
	f.String("builder", b.Builder, "")
	f.StringArray("cache-from", b.CacheFrom, "")
	f.StringArray("cache-to", b.CacheTo, "")
	f.String("call", b.Call, "")
	f.String("cgroup-parent", b.CGroupParent, "")
	f.Bool("check", b.Check, "")
	f.Bool("debug", b.Debug, "")
	f.String("file", b.File, "")
	f.String("iidfile", b.IIDFile, "")
	f.StringArray("label", b.Label, "")
	f.Bool("load", b.Load, "")
	f.String("metadata-file", b.MetadataFile, "")
	f.String("network", b.Network, "")
	f.Bool("no-cache", b.NoCache, "")
	f.StringArray("no-cache-filter", b.NoCacheFilter, "")
	f.StringArray("output", b.Output, "")
	f.StringArray("platform", b.Platform, "")
	f.String("progress", b.Progress, "")
	f.Bool("provenance", b.Provenance, "")
	f.Bool("pull", b.Pull, "")
	f.Bool("push", b.Push, "")
	f.Bool("quiet", b.Quiet, "")
	f.Bool("sbom", b.Sbom, "")
	f.StringArray("secret", b.Secret, "")
	f.String("shm-size", b.ShmSize, "")
	f.StringArray("ssh", b.SSH, "")
	f.StringArray("tag", b.Tag, "")
	f.String("target", b.Target, "")
	f.String("ulimit", b.ULimit, "")

	if err := f.Parse(cleanArgs); err != nil {
		return "", err
	}

	pterm.Debug.Printfln("running docker build for %q", b.Context)

	return util.NewDockerCommand().Build(b.Context).
		TemplateData(map[string]any{"Squadron": squadron, "Unit": unit, "Build": b}).
		Args(qflag.Parse(f)...).
		Run(ctx)
}
