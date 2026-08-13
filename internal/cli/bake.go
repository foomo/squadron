package cli

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
				pterm.Info.Printfln("💾 | writing output to %s", output)

				if err := os.WriteFile(output, bakefile, 0600); err != nil {
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

	addPushFlag(flags, x)
	addParallelFlag(flags, x)
	addOutputFlag(flags, x)
	addBakeArgsFlag(flags, x)
	addPushArgsFlag(flags, x)
	addTagsFlag(flags, x)

	return cmd
}
