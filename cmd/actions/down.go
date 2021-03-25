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
		Use:     "down",
		Short:   "uninstalls the squadron chart",
		Example: "  squadron down --namespace demo",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, helmArgs := parseExtraArgs(args)
			return down(log, cwd, flagNamespace, helmArgs)
		},
	}
)

func down(l *logrus.Entry, cwd, namespace string, extraArgs []string) error {
	sq, err := squadron.New(l, cwd, namespace, nil)
	if err != nil {
		return err
	}
	return sq.Down(extraArgs)
}
