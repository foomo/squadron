package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	statusCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
}

var statusCmd = &cobra.Command{
	Use:     "status [UNIT...]",
	Short:   "installs the squadron or given units",
	Example: "  squadron status frontend backend --namespace demo --build --push -- --dry-run",
	RunE: func(cmd *cobra.Command, args []string) error {
		return status(args, cwd, flagNamespace, flagFiles)
	},
}

func status(args []string, cwd, namespace string, files []string) error {
	sq := squadron.New(cwd, namespace, files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)

	unitsNames, err := parseUnitNames(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	if unitsNames != nil {
		if err := sq.FilterConfig(unitsNames); err != nil {
			return err
		}
	}

	if err := sq.RenderConfig(); err != nil {
		return err
	}

	units, err := parseUnitArgs(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	return sq.Status(units, helmArgs)
}
