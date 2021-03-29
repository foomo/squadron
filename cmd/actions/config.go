package actions

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

var (
	configCmd = &cobra.Command{
		Use:     "config",
		Short:   "generate and view the squadron config",
		Example: "  squadron config --file squadron.yaml --file squadron.override.yaml",
		Args:    cobra.MinimumNArgs(0),
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

	cf, err := sq.GetConfigYAML()
	if err != nil {
		return err
	}

	fmt.Println(string(cf))
	return nil
}
