package actions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	configCmd.Flags().BoolVar(&flagNoRender, "no-render", false, "don't render the config template")
}

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "generate and view the squadron config",
	Example: "  squadron config --file squadron.yaml --file squadron.override.yaml",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return config(cwd, flagFiles, flagNoRender)
	},
}

func config(cwd string, files []string, noRender bool) error {
	sq := squadron.New(cwd, "", files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	if !noRender {
		if err := sq.RenderConfig(); err != nil {
			return err
		}
	}

	fmt.Println(sq.GetConfigYAML())
	return nil
}
