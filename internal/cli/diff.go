package cli

import (
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDiff(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "diff [SQUADRON] [UNIT...]",
		Short:   "shows the diff between the installed and local chart",
		Example: "  squadron diff storefinder frontend backend --namespace demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, x.GetString("namespace"), c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			args, helmArgs := parseExtraArgs(args)

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if err := sq.UpdateLocalDependencies(cmd.Context(), x.GetInt("parallel")); err != nil {
				return errors.Wrap(err, "failed to update dependencies")
			}

			out, err := sq.Diff(cmd.Context(), helmArgs, x.GetInt("parallel"))
			if err != nil {
				return err
			}

			if !x.GetBool("raw") {
				out = util.Highlight(out)
			}

			pterm.Println(out)

			return nil
		},
	}

	flags := cmd.Flags()

	addNamespaceFlag(flags, x)
	addParallelFlag(flags, x)
	addTagsFlag(flags, x)
	addRawFlag(flags, x)

	return cmd
}
