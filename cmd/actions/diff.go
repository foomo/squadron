package actions

import (
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewDiff(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "diff [SQUADRON] [UNIT...]",
		Short:   "shows the diff between the installed and local chart",
		Example: "  squadron diff storefinder frontend backend --namespace demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, c.GetString("namespace"), c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			args, helmArgs := parseExtraArgs(args)

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, c.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if err := sq.UpdateLocalDependencies(cmd.Context(), c.GetInt("parallel")); err != nil {
				return errors.Wrap(err, "failed to update dependencies")
			}

			out, err := sq.Diff(cmd.Context(), helmArgs, c.GetInt("parallel"))
			if err != nil {
				return err
			}

			if !c.GetBool("raw") {
				out = util.Highlight(out)
			}
			pterm.Println(out)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringP("namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	_ = c.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.Int("parallel", 1, "run command in parallel")
	_ = c.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = c.BindPFlag("tags", flags.Lookup("tags"))

	flags.Bool("raw", false, "print raw output without highlighting")
	_ = c.BindPFlag("raw", flags.Lookup("raw"))

	return cmd
}
