package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var generateCmd = &cobra.Command{
	Use:     "generate",
	Short:   "generate and view the squadron chart",
	Example: "  squadron generate fronted backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return generate(cwd, flagFiles)
	},
}

func generate(cwd string, files []string) error {
	sq := squadron.New(cwd, "", files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	if err := sq.RenderConfig(); err != nil {
		return err
	}

	return sq.Generate(sq.GetConfig().Units)
}
