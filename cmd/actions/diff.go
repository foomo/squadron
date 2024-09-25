package actions

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	diffCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	diffCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	diffCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var diffCmd = &cobra.Command{
	Use:     "diff [SQUADRON] [UNIT...]",
	Short:   "shows the diff between the installed and local chart",
	Example: "  squadron diff storefinder frontend backend --namespace demo",
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

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to render config")
		}

		if err := sq.UpdateLocalDependencies(cmd.Context(), flagParallel); err != nil {
			return errors.Wrap(err, "failed to update dependencies")
		}

		return sq.Diff(cmd.Context(), helmArgs, flagParallel)
	},
}
