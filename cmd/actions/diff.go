package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	diffCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
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

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames, flagTags); err != nil {
			return err
		}

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return err
		}

		return sq.Diff(cmd.Context(), helmArgs, flagParallel)
	},
}
