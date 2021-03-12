package actions

import (
	"github.com/spf13/cobra"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	upCmd.Flags().BoolVarP(&flagNoBuild, "no-build", "nb", false, "Build service squadron before publishing")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the built service to the registry")
}

var (
	flagNamespace string
	flagNoBuild   bool
	flagPush      bool
)

var (
	upCmd = &cobra.Command{
		Use:   "up [SQUADRON] -n {NAMESPACE}",
		Short: "builds and installs a group of charts",
		Long:  "builds and installs a group of services with given namespace",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// todo if one service specified dont use all units
			units := sq.Units()
			if !flagNoBuild {
				// todo build
			}
			if flagPush {
				// todo push
			}
			// todo check what else args will contain
			return sq.Up(units, flagNamespace, args...)
		},
	}
)
