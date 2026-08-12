package cli

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// addBakeFlag registers the shared `--bake` flag and binds it to the given viper instance
func addBakeFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.Bool("bake", false, "bakes or rebakes units")
	_ = x.BindPFlag("bake", flags.Lookup("bake"))
}

// addBakeArgsFlag registers the shared `--bake-args` flag and binds it to the given viper instance
func addBakeArgsFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.StringArray("bake-args", nil, "additional docker buildx bake args")
	_ = x.BindPFlag("bake-args", flags.Lookup("bake-args"))
}

// addBuildFlag registers the shared `--build` flag and binds it to the given viper instance
func addBuildFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.Bool("build", false, "builds or rebuilds units")
	_ = x.BindPFlag("build", flags.Lookup("build"))
}

// addBuildArgsFlag registers the shared `--build-args` flag and binds it to the given viper instance
func addBuildArgsFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.StringArray("build-args", nil, "additional docker buildx build args")
	_ = x.BindPFlag("build-args", flags.Lookup("build-args"))
}

// addNamespaceFlag registers the shared `--namespace` flag and binds it to the given viper instance
func addNamespaceFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.StringP("namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	_ = x.BindPFlag("namespace", flags.Lookup("namespace"))
}

// addOutputFlag registers the shared `--output` flag and binds it to the given viper instance
func addOutputFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.String("output", "", "write the output to the given path")
	_ = x.BindPFlag("output", flags.Lookup("output"))
}

// addParallelFlag registers the shared `--parallel` flag and binds it to the given viper instance
func addParallelFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.Int("parallel", 1, "run command in parallel")
	_ = x.BindPFlag("parallel", flags.Lookup("parallel"))
}

// addPushFlag registers the shared `--push` flag and binds it to the given viper instance
func addPushFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.Bool("push", false, "pushes built squadron units to the registry")
	_ = x.BindPFlag("push", flags.Lookup("push"))
}

// addPushArgsFlag registers the shared `--push-args` flag and binds it to the given viper instance
func addPushArgsFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.StringArray("push-args", nil, "additional docker push args")
	_ = x.BindPFlag("push-args", flags.Lookup("push-args"))
}

// addRawFlag registers the shared `--raw` flag and binds it to the given viper instance
func addRawFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.Bool("raw", false, "print raw output without highlighting")
	_ = x.BindPFlag("raw", flags.Lookup("raw"))
}

// addTagsFlag registers the shared `--tags` flag and binds it to the given viper instance
func addTagsFlag(flags *pflag.FlagSet, x *viper.Viper) {
	flags.StringSlice("tags", nil, "list of tags to include (any match) or exclude with a '-' prefix (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = x.BindPFlag("tags", flags.Lookup("tags"))
}
