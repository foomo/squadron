package actions

import (
	"fmt"

	"github.com/foomo/squadron/internal/util"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	configCmd.Flags().BoolVar(&flagNoRender, "no-render", false, "don't render the config template")
}

var configCmd = &cobra.Command{
	Use:     "config [SQUADRON] [UNIT...]",
	Short:   "generate and view the squadron config",
	Example: "  squadron config storefinder frontend backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, "", flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return err
		}

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames); err != nil {
			return err
		}

		if !flagNoRender {
			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return err
			}
		}

		fmt.Print(util.Highlight(sq.ConfigYAML()))

		return nil
	},
}
