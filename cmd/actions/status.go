package actions

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	statusCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	statusCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	statusCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var statusCmd = &cobra.Command{
	Use:     "status [SQUADRON] [UNIT...]",
	Short:   "installs the squadron or given units",
	Example: "  squadron status storefinder frontend backend --namespace demo",
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, flagNamespace, flagFiles)

		if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to merge config files")
		}

		args, helmArgs := parseExtraArgs(args)

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, flagTags); err != nil {
			return errors.Wrap(err, "failed to filter config")
		}

		return sq.Status(cmd.Context(), helmArgs, flagParallel)
	},
}
