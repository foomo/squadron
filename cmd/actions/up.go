package actions

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	upCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes units to the registry")
}

var (
	upCmd = &cobra.Command{
		Use:     "up [UNIT...]",
		Short:   "installs the squadron chart",
		Example: "  squadron up frontend backend --build --push --namespace demo -- --dry-run",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			unitNames, helmArgs := parseExtraArgs(args)
			return up(log, unitNames, cwd, flagNamespace, flagBuild, flagPush, flagFiles, helmArgs)
		},
	}
)

func up(l *logrus.Entry, unitNames []string, cwd, namespace string, build, push bool, files []string, helmArgs []string) error {
	sq, err := squadron.New(l, cwd, namespace, files)
	if err != nil {
		return err
	}

	units := map[string]squadron.Unit{}
	if len(unitNames) == 0 {
		units = sq.Units()
	}
	for _, un := range unitNames {
		units[un] = sq.Units()[un]
	}

	for _, unit := range units {
		if build {
			if err := sq.Build(unit); err != nil {
				return err
			}
		}
		if push {
			if err := sq.Push(unit); err != nil {
				return err
			}
		}
	}
	return sq.Up(units, namespace, helmArgs)
}
