package actions

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes built squadron units to the registry")
}

var (
	buildCmd = &cobra.Command{
		Use:     "build [UNIT...]",
		Short:   "build or rebuild squadron units",
		Example: "  squadron build frontend backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return build(log, args, cwd, flagFiles, flagPush)
		},
	}
)

func build(l *logrus.Entry, unitNames []string, cwd string, files []string, push bool) error {
	sq, err := squadron.New(l, cwd, "", files)
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
		if err := sq.Build(unit); err != nil {
			return err
		}
		if push {
			if err := sq.Push(unit); err != nil {
				return err
			}
		}
	}
	return nil
}
