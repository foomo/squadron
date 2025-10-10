package actions

import (
	"context"
	"fmt"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/config"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewList(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "list [SQUADRON]",
		Short:   "list squadron units",
		Example: "  squadron list storefinder",
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

			var list pterm.LeveledList

			// List squadrons
			_ = sq.Config().Squadrons.Iterate(cmd.Context(), func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
				list = append(list, pterm.LeveledListItem{Level: 0, Text: key})

				return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
					list = append(list, pterm.LeveledListItem{Level: 1, Text: k})
					if x.GetBool("with-tags") && len(v.Tags) > 0 {
						list = append(list, pterm.LeveledListItem{Level: 2, Text: "🏷️: " + v.Tags.SortedString()})
					}

					if x.GetBool("with-charts") && len(v.Chart.String()) > 0 {
						list = append(list, pterm.LeveledListItem{Level: 2, Text: "📑: " + v.Chart.String()})
					}

					if x.GetBool("with-priority") && len(v.Chart.String()) > 0 {
						list = append(list, pterm.LeveledListItem{Level: 2, Text: fmt.Sprintf("☝️: %d", v.Priority)})
					}

					if x.GetBool("with-bakes") && len(v.Bakes) > 0 {
						for name, build := range v.Bakes {
							list = append(list, pterm.LeveledListItem{Level: 2, Text: "📦: " + name})
							for _, tag := range build.Tags {
								list = append(list, pterm.LeveledListItem{Level: 3, Text: tag})
							}
						}
					}

					if x.GetBool("with-builds") && len(v.Builds) > 0 {
						for name, build := range v.Builds {
							list = append(list, pterm.LeveledListItem{Level: 2, Text: "📦: " + name})

							list = append(list, pterm.LeveledListItem{Level: 3, Text: build.Image + ":" + build.Tag})
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

	flags := cmd.Flags()

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = x.BindPFlag("tags", flags.Lookup("tags"))

	flags.Bool("with-tags", false, "include tags")
	_ = x.BindPFlag("with-tags", flags.Lookup("with-tags"))

	flags.Bool("with-charts", false, "include charts")
	_ = x.BindPFlag("with-charts", flags.Lookup("with-charts"))

	flags.Bool("with-priority", false, "include priority")
	_ = x.BindPFlag("with-priority", flags.Lookup("with-priority"))

	flags.Bool("with-builds", false, "include builds")
	_ = x.BindPFlag("with-builds", flags.Lookup("with-builds"))

	flags.Bool("with-bakes", false, "include bakes")
	_ = x.BindPFlag("with-bakes", flags.Lookup("with-bakes"))

	return cmd
}
