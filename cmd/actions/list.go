package actions

import (
	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/config"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
)

var (
	flagBuilds bool
	flagCharts bool
)

func init() {
	listCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	listCmd.Flags().BoolVar(&flagCharts, "charts", false, "include charts")
	listCmd.Flags().BoolVar(&flagBuilds, "builds", false, "include builds")
}

var listCmd = &cobra.Command{
	Use:     "list [SQUADRON]",
	Short:   "list squadron units",
	Example: "  squadron list storefinder",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, "", flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return err
		}

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames, flagTags); err != nil {
			return err
		}

		var list pterm.LeveledList

		// List squadrons
		_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
			list = append(list, pterm.LeveledListItem{Level: 0, Text: key})
			return value.Iterate(func(k string, v *config.Unit) error {
				list = append(list, pterm.LeveledListItem{Level: 1, Text: k})
				if flagCharts {
					list = append(list, pterm.LeveledListItem{Level: 2, Text: "📑: " + v.Chart.String()})
				}
				if flagBuilds {
					for name, build := range v.Builds {
						list = append(list, pterm.LeveledListItem{Level: 2, Text: "📦: " + name})
						for _, dependency := range build.Dependencies {
							list = append(list, pterm.LeveledListItem{Level: 3, Text: "🗃️: " + dependency})
						}
					}
				}
				return nil
			})
		})

		if len(list) > 0 {
			root := putils.TreeFromLeveledList(list)
			root.Text = "Squadron"
			return pterm.DefaultTree.WithRoot(root).Render()
		}

		return nil
	},
}
