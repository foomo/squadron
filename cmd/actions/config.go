package actions

import (
	"github.com/foomo/squadron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	configCmd.Flags().StringSliceVarP(&flagFiles, "file", "f", []string{}, "specify alternative squadron files (default squadron.yaml)")
}

var (
	configCmd = &cobra.Command{
		Use:   "config",
		Short: "validate and view the Squadron file.",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return config(log, cwd, flagFiles)
		},
	}
)

func config(l *logrus.Entry, cwd string, files []string) error {
	sq, err := squadron.New(l, cwd, "", files)
	if err != nil {
		return err
	}
	return sq.Config()
}
