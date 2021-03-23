package actions

import (
	"github.com/foomo/squadron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the built service to the registry")
}

var (
	buildCmd = &cobra.Command{
		Use:   "build {UNIT...} -p",
		Short: "builds and pushes a unit image",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return build(log, args, cwd, flagPush)
		},
	}
)

func build(l *logrus.Entry, unitNames []string, cwd string, push bool) error {
	sq, err := squadron.New(l, cwd, "")
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
