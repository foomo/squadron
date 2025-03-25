package actions

import (
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTemplate(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template [SQUADRON] [UNIT...]",
		Short:   "render chart templates locally and display the output",
		Example: "  squadron template storefinder frontend backend --namespace demo",
		Args:    cobra.MinimumNArgs(0),
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

			out, err := sq.Template(cmd.Context(), helmArgs, c.GetInt("parallel"))
			if err != nil {
				return errors.Wrap(err, "failed to render template")
			}

			pterm.Println(util.Highlight(out))

			return nil
		},
	}

	flags := cmd.Flags()
	flags.Int("parallel", 1, "run command in parallel")
	_ = c.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.StringP("namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	_ = c.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = c.BindPFlag("tags", flags.Lookup("tags"))

	return cmd
}
