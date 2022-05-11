package actions

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	rollbackCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	rollbackCmd.Flags().StringVarP(&flagRevision, "revision", "r", "", "specifies the revision to roll back to")
}

var rollbackCmd = &cobra.Command{
	Use:     "rollback [UNIT...]",
	Short:   "rolls back the squadron or given units",
	Example: "  squadron rollback frontend backend --namespace demo",
	RunE: func(cmd *cobra.Command, args []string) error {
		return rollback(cmd.Context(), args, cwd, flagNamespace, flagRevision, flagFiles)
	},
}

func rollback(ctx context.Context, args []string, cwd, namespace string, revision string, files []string) error {
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

	if err := sq.RenderConfig(ctx); err != nil {
		return err
	}

	units, err := parseUnitArgs(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	return sq.Rollback(ctx, units, revision, helmArgs)
}
