package actions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var flagPrefixSquadron bool

func init() {
	listCmd.Flags().BoolVar(&flagPrefixSquadron, "prefix-squadron", false, "add squadron prefix")
}

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "list squadron units",
	Example: "  squadron list",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return list(cwd, flagFiles)
	},
}

func list(cwd string, files []string) error {
	sq := squadron.New(cwd, "", files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	_ = sq.GetConfig().Units.Iterate(func(name string, unit *squadron.Unit) error {
		if flagPrefixSquadron {
			fmt.Printf("%s/%s\n", sq.Name(), name)
		} else {
			fmt.Println(name)
		}
		return nil
	})

	return nil
}
