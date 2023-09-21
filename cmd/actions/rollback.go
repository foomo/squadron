package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	rollbackCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	rollbackCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	rollbackCmd.Flags().StringVarP(&flagRevision, "revision", "r", "", "specifies the revision to roll back to")
}

var rollbackCmd = &cobra.Command{
	Use:     "rollback [SQUADRON] [UNIT...]",
	Short:   "rolls back the squadron or given units",
	Example: "  squadron rollback storefinder frontend backend --namespace demo",
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

		return sq.Rollback(cmd.Context(), flagRevision, helmArgs, flagParallel)
	},
}
