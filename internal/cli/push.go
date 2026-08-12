package cli

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPush(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "push [SQUADRON] [UNIT...]",
		Short:   "pushes the squadron or given units",
		Example: "  squadron push storefinder frontend backend --namespace demo --build",
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, x.GetString("namespace"), c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if x.GetBool("bake") {
				bakefile, err := sq.Bakefile(cmd.Context())
				if err != nil {
					return errors.Wrap(err, "failed to bake units")
				}

				if err := sq.Bake(cmd.Context(), bakefile, x.GetStringSlice("bake-args")); err != nil {
					return errors.Wrap(err, "failed to bake units")
				}
			}

			if x.GetBool("build") {
				if err := sq.Build(cmd.Context(), x.GetStringSlice("build-args"), x.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to build units")
				}
			}

			return sq.Push(cmd.Context(), x.GetStringSlice("push-args"), x.GetInt("parallel"))
		},
	}

	flags := cmd.Flags()

	addNamespaceFlag(flags, x)
	addBuildFlag(flags, x)
	addBakeFlag(flags, x)
	addParallelFlag(flags, x)
	addBakeArgsFlag(flags, x)
	addBuildArgsFlag(flags, x)
	addPushArgsFlag(flags, x)
	addTagsFlag(flags, x)

	return cmd
}
