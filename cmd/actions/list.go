package actions

import (
	"fmt"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/config"
	"github.com/spf13/cobra"
)

var (
	flagBuilds bool
)

func init() {
	listCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
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

		// List squadrons
		_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
			fmt.Println("Squadron:", key)
			return value.Iterate(func(k string, v *config.Unit) error {
				fmt.Println("  " + k)
				if flagBuilds {
					for name, build := range v.Builds {
						fmt.Println("    " + name)
						for _, dependency := range build.Dependencies {
							fmt.Println("      " + dependency)
						}
					}
				}
				return nil
			})
		})

		return nil
	},
}
