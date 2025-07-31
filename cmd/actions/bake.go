package actions

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewBake(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bake [SQUADRON.UNIT...]",
		Short:   "bake or rebake squadron units",
		Example: "squadron bake storefinder frontend backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, "", c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, c.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if err := sq.Bake(cmd.Context(), c.GetStringSlice("bake-args")); err != nil {
				return errors.Wrap(err, "failed to bake units")
			}

			if c.GetBool("push") {
				if err := sq.Push(cmd.Context(), c.GetStringSlice("push-args"), c.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to push units")
				}
			}

			return nil
		},
	}

	flags := cmd.Flags()
	flags.BoolP("push", "p", false, "pushes built squadron units to the registry")
	_ = c.BindPFlag("push", flags.Lookup("push"))

	cmd.Flags().Int("parallel", 1, "run command in parallel")
	_ = c.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.StringArray("bake-args", nil, "additional docker bake args")
	_ = c.BindPFlag("bake-args", flags.Lookup("bake-args"))

	flags.StringArray("push-args", nil, "additional docker push args")
	_ = c.BindPFlag("push-args", flags.Lookup("push-args"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = c.BindPFlag("tags", flags.Lookup("tags"))

	return cmd
}
