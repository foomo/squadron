package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	rollbackCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	rollbackCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	rollbackCmd.Flags().StringVarP(&flagRevision, "revision", "r", "", "specifies the revision to roll back to")
	rollbackCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
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
		if err := sq.FilterConfig(squadronName, unitNames, flagTags); err != nil {
			return err
		}

		return sq.Rollback(cmd.Context(), flagRevision, helmArgs, flagParallel)
	},
}
