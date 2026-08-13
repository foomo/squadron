package cli

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewBuild(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "build [SQUADRON.UNIT...]",
		Short:   "build or rebuild squadron units",
		Example: "squadron build storefinder frontend backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, "", c.GetStringSlice("file"))

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

			if err := sq.Build(cmd.Context(), x.GetStringSlice("build-args"), x.GetInt("parallel")); err != nil {
				return errors.Wrap(err, "failed to build units")
			}

			if x.GetBool("push") {
				if err := sq.Push(cmd.Context(), x.GetStringSlice("push-args"), x.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to push units")
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()

	addPushFlag(flags, x)
	addParallelFlag(flags, x)
	addBuildArgsFlag(flags, x)
	addPushArgsFlag(flags, x)
	addTagsFlag(flags, x)

	return cmd
}
