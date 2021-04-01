package actions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	templateCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
}

var (
	templateCmd = &cobra.Command{
		Use:     "template",
		Short:   "render chart templates locally and display the output",
		Example: "  squadron template",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			return template(args, cwd, flagNamespace, flagFiles)
		},
	}
)

func template(args []string, cwd, namespace string, files []string) error {
	sq, err := squadron.New(cwd, namespace, files)
	if err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)

	if err := sq.Generate(sq.GetUnits()); err != nil {
		return err
	} else if out, err := sq.Template(helmArgs); err != nil {
		return err
	} else {
		fmt.Println(out)
	}

	return nil
}
