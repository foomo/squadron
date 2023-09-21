package actions

import (
	"fmt"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	templateCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	templateCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var templateCmd = &cobra.Command{
	Use:     "template [SQUADRON] [UNIT...]",
	Short:   "render chart templates locally and display the output",
	Example: "  squadron template storefinder frontend backend --namespace demo",
	Args:    cobra.MinimumNArgs(0),
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

		out, err := sq.Template(cmd.Context(), helmArgs)
		if err != nil {
			return err
		}

		fmt.Print(util.Highlight(out))

		return nil
	},
}
