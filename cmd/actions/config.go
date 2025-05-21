package actions

import (
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfig(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "config [SQUADRON] [UNIT...]",
		Short:   "generate and view the squadron config",
		Example: "  squadron config storefinder frontend backend",
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

			if !c.GetBool("no-render") {
				if err := sq.RenderConfig(cmd.Context()); err != nil {
					return errors.Wrap(err, "failed to render config")
				}
			}

			out := sq.ConfigYAML()
			if !c.GetBool("raw") {
				out = util.Highlight(out)
			}
			pterm.Println(out)

			return nil
		},
	}

	flags := cmd.Flags()
	flags.Bool("no-render", false, "don't render the config template")
	_ = c.BindPFlag("no-render", flags.Lookup("no-render"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = c.BindPFlag("tags", flags.Lookup("tags"))

	flags.Bool("raw", false, "print raw output without highlighting")
	_ = c.BindPFlag("raw", flags.Lookup("raw"))

	return cmd
}
