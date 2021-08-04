package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var generateCmd = &cobra.Command{
	Use:     "generate [UNIT...]",
	Short:   "generate and view the squadron or given units charts",
	Example: "  squadron generate fronted backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return generate(args, cwd, flagFiles)
	},
}

func generate(args []string, cwd string, files []string) error {
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

	if err := sq.RenderConfig(); err != nil {
		return err
	}

	return sq.Generate(sq.GetConfig().Units)
}
