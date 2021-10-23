package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes built squadron units to the registry")
}

var buildCmd = &cobra.Command{
	Use:     "build [UNIT...]",
	Short:   "build or rebuild squadron units",
	Example: "  squadron build frontend backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return build(cmd.Context(), args, cwd, flagFiles, flagPush)
	},
}

func build(ctx context.Context, args []string, cwd string, files []string, push bool) error {
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

	if err := sq.RenderConfig(ctx); err != nil {
		return err
	}

	units, err := parseUnitArgs(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	for _, unit := range units {
		if err := unit.Build(ctx); err != nil {
			return err
		}
	}

	if push {
		for _, unit := range units {
			if err := unit.Push(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}
