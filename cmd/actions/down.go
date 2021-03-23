package actions

import (
	"github.com/foomo/squadron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	downCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	downCmd = &cobra.Command{
		Use:   "down -n {NAMESPACE}",
		Short: "uninstalls a group of charts",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			var extraArgs []string
			var units []string
			for i, arg := range args {
				if arg == "--" {
					extraArgs = args[i:]
					units = args[:i]
				}
			}
			return down(log, units, cwd, flagNamespace, extraArgs)
		},
	}
)

func down(l *logrus.Entry, unitNames []string, cwd, namespace string, extraArgs []string) error {
	sq, err := squadron.New(l, cwd, namespace, nil)
	if err != nil {
		return err
	}

	var units map[string]squadron.Unit
	if len(unitNames) == 0 {
		units = sq.Units()
	}
	for _, un := range unitNames {
		units[un] = sq.Units()[un]
	}
	return sq.Down(extraArgs)
}
