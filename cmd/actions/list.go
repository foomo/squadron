package actions

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list squadron units",
	Example: "  squadron list",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return list(cmd.Context(), cwd, flagFiles)
	},
}

func list(ctx context.Context, cwd string, files []string) error {
	sq := squadron.New(cwd, "", files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	for name := range sq.GetConfig().Units {
		fmt.Println(name)
	}

	return nil
}
