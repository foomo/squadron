package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	upCmd.Flags().StringVar(&flagUnit, "unit", "", "Specifies the unit")
	upCmd.Flags().BoolVar(&flagNoBuild, "no-build", false, "Build service squadron before publishing")
	upCmd.Flags().BoolVar(&flagPush, "push", false, "Pushes the built service to the registry")
}

var (
	flagNamespace string
	flagNoBuild   bool
	flagPush      bool
	flagUnit      string
)

var (
	upCmd = &cobra.Command{
		Use:   "up [SQUADRON] -n {NAMESPACE}",
		Short: "builds and installs a group of charts",
		Long:  "builds and installs a group of services with given namespace",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			units := sq.Units()
			if flagUnit != "" {
				for name, unit := range units {
					if name == flagUnit {
						units = map[string]squadron.Unit{name: unit}
					}
				}
			}
			for _, unit := range units {
				if !flagNoBuild {
					if err := sq.Build(unit); err != nil {
						return err
					}
				}
				if flagPush {
					if err := sq.Push(unit); err != nil {
						return err
					}
				}
			}
			// todo check what else args will contain
			return sq.Up(units, flagNamespace, args...)
		},
	}
)
