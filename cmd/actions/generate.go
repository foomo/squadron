package actions

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var (
	generateCmd = &cobra.Command{
		Use:     "generate",
		Short:   "generate and view the squadron chart",
		Example: "  squadron generate fronted backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(log, cwd, flagFiles)
		},
	}
)

func generate(l *logrus.Entry, cwd string, files []string) error {
	sq, err := squadron.New(l, cwd, "", files)
	if err != nil {
		return err
	}

	return sq.Generate(sq.GetUnits())
}
