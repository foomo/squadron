package actions

import (
	"os"

	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewBake(c *viper.Viper) *cobra.Command {
	x := viper.New()

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
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			bakefile, err := sq.Bakefile(cmd.Context())
			if err != nil {
				return errors.Wrap(err, "failed to generate bakefile")
			}

			if output := x.GetString("output"); output != "" {
				hcl, err := bakefile.HCL()
				if err != nil {
					return errors.Wrap(err, "failed to marshal bake config")
				}

				pterm.Info.Printfln("💾 | writing output to %s", output)
				if err := os.WriteFile(output, hcl, 0600); err != nil {
					return errors.Wrap(err, "failed to write bakefile")
				}

				return nil
			}

			if err := sq.Bake(cmd.Context(), bakefile, x.GetStringSlice("bake-args")); err != nil {
				return errors.Wrap(err, "failed to bake units")
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
	flags.Bool("push", false, "pushes built squadron units to the registry")
	_ = x.BindPFlag("push", flags.Lookup("push"))

	cmd.Flags().Int("parallel", 1, "run command in parallel")
	_ = x.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.String("output", "", "write the output to the given path")
	_ = x.BindPFlag("output", flags.Lookup("output"))

	flags.StringArray("bake-args", nil, "additional docker bake args")
	_ = x.BindPFlag("bake-args", flags.Lookup("bake-args"))

	flags.StringArray("push-args", nil, "additional docker push args")
	_ = x.BindPFlag("push-args", flags.Lookup("push-args"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = x.BindPFlag("tags", flags.Lookup("tags"))

	return cmd
}
