package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
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
			return down(args, cwd, flagNamespace)
		},
	}
)

func down(args []string, cwd, namespace string) error {
	sq, err := squadron.New(cwd, namespace, nil)
	if err != nil {
		return err
	}

	_, helmArgs := parseExtraArgs(args)

	return sq.Down(helmArgs)
}
