package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	pushCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	pushCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
}

var pushCmd = &cobra.Command{
	Use:     "push [UNIT...]",
	Short:   "pushes the squadron or given units",
	Example: "  squadron push frontend backend --namespace demo --build",
	RunE: func(cmd *cobra.Command, args []string) error {
		return push(args, cwd, flagNamespace, flagBuild, flagFiles)
	},
}

func push(args []string, cwd, namespace string, build bool, files []string) error {
	sq := squadron.New(cwd, namespace, files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

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

	if build {
		for _, unit := range units {
			if err := unit.Build(); err != nil {
				return err
			}
		}
	}

	for _, unit := range units {
		if err := unit.Push(); err != nil {
			return err
		}
	}

	return nil
}
