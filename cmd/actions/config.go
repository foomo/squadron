package actions

import (
	"fmt"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.Flags().BoolVar(&flagNoRender, "no-render", false, "don't render the config template")
	configCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var configCmd = &cobra.Command{
	Use:     "config [SQUADRON] [UNIT...]",
	Short:   "generate and view the squadron config",
	Example: "  squadron config storefinder frontend backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, "", flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return errors.Wrap(err, "failed to merge config files")
		}

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames, flagTags); err != nil {
			return errors.Wrap(err, "failed to filter config")
		}

		if !flagNoRender {
			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}
		}

		fmt.Print(util.Highlight(sq.ConfigYAML()))

		return nil
	},
}
