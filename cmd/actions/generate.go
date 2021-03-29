package actions

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var (
	generateCmd = &cobra.Command{
		Use:     "generate [UNIT...]",
		Short:   "generate and view the squadron chart",
		Example: "  squadron generate fronted backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(log, args, cwd, flagFiles)
		},
	}
)

func generate(l *logrus.Entry, args []string, cwd string, files []string) error {
	sq, err := squadron.New(l, cwd, "", files)
	if err != nil {
		return err
	}

	units, err := parseUnitArgs(args, sq.GetUnits())
	if err != nil {
		return err
	}

	return sq.Generate(units)
}
