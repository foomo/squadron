package actions

import (
	"path"

	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [NAME]",
	Short: "initializes an example",
	Long:  "initializes an example project with squadron",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := initialize(args[0], flagDir, flagVerbose)
		return err
	},
}

func initialize(name, dir string, flagVerbose bool) (string, error) {
	return squadron.Init(log, path.Join(dir, name))
}
