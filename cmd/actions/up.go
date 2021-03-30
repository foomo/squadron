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
		Example: "  squadron up frontend backend --namespace demo --build --push -- --dry-run",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return up(log, args, cwd, flagNamespace, flagBuild, flagPush, flagFiles)
		},
	}
)

func up(l *logrus.Entry, args []string, cwd, namespace string, build, push bool, files []string) error {
	sq, err := squadron.New(l, cwd, namespace, files)
	if err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)
	units, err := parseUnitArgs(args, sq.GetUnits())
	if err != nil {
		return err
	}

	for _, unit := range units {
		if err := unit.Build(build); err != nil {
			return err
		}
	}

	if push {
		for _, unit := range units {
			if err := unit.Push(); err != nil {
				return err
			}
		}
	}

	if err := sq.Generate(units); err != nil {
		return err
	} else if err := sq.Up(helmArgs); err != nil {
		return err
	}

	return nil
}
