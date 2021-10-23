package actions

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	configCmd.Flags().BoolVar(&flagNoRender, "no-render", false, "don't render the config template")
}

var configCmd = &cobra.Command{
	Use:     "config [UNIT...]",
	Short:   "generate and view the squadron config",
	Example: "  squadron config --file squadron.yaml --file squadron.override.yaml",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config(cmd.Context(), args, cwd, flagFiles, flagNoRender)
	},
}

func config(ctx context.Context, args []string, cwd string, files []string, noRender bool) error {
	sq := squadron.New(cwd, "", files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	unitsNames, err := parseUnitNames(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	if unitsNames != nil {
		if err := sq.FilterConfig(unitsNames); err != nil {
			return err
		}
	}

	if !noRender {
		if err := sq.RenderConfig(ctx); err != nil {
			return err
		}
	}

	fmt.Println(sq.GetConfigYAML())
	return nil
}
