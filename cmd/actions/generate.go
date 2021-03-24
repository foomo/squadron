package actions

import (
	"github.com/foomo/squadron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	generateCmd.Flags().StringSliceVarP(&flagFiles, "file", "f", []string{}, "Configuration file to merge")
}

var (
	generateCmd = &cobra.Command{
		Use:   "generate {UNIT...} -f",
		Short: "builds and installs a group of charts",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return generate(log, args, cwd, flagFiles)
		},
	}
)

func generate(l *logrus.Entry, unitNames []string, cwd string, files []string) error {
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
	return sq.Generate(units)
}
