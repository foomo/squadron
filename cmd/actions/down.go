package actions

import (
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
)

func init() {
	downCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	downCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var downCmd = &cobra.Command{
	Use:     "down [SQUADRON] [UNIT...]",
	Short:   "uninstalls the squadron or given units",
	Example: "  squadron down storefinder frontend backend --namespace demo",
	Args:    cobra.MinimumNArgs(0),
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

		return sq.Down(cmd.Context(), helmArgs, flagParallel)
	},
}
