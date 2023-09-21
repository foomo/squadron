package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	statusCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	statusCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
}

var statusCmd = &cobra.Command{
	Use:     "status [SQUADRON] [UNIT...]",
	Short:   "installs the squadron or given units",
	Example: "  squadron status storefinder frontend backend --namespace demo",
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, flagNamespace, flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return err
		}

		args, helmArgs := parseExtraArgs(args)

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames); err != nil {
			return err
		}

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return err
		}

		return sq.Status(cmd.Context(), helmArgs, flagParallel)
	},
}
