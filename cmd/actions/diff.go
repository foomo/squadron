package actions

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	diffCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	diffCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
}

var diffCmd = &cobra.Command{
	Use:     "diff [SQUADRON] [UNIT...]",
	Short:   "shows the diff between the installed and local chart",
	Example: "  squadron diff storefinder frontend backend --namespace demo",
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, flagNamespace, flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return err
		}

		args, helmArgs := parseExtraArgs(args)

		if len(args) > 0 {
			if err := sq.Config().Squadrons.Filter(args[0]); err != nil {
				return errors.Wrap(err, "invalid SQUADRON argument")
			}
		}

		if len(args) > 1 {
			if err := sq.Config().Squadrons[args[0]].Filter(args[1:]...); err != nil {
				return errors.Wrap(err, "invalid UNIT argument")
			}
		}

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return err
		}

		return sq.Diff(cmd.Context(), helmArgs, flagParallel)
	},
}
