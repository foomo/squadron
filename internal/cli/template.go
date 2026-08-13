package cli

import (
	"os"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewTemplate(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "template [SQUADRON] [UNIT...]",
		Short:   "render chart templates locally and display the output",
		Example: "  squadron template storefinder frontend backend --namespace demo",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, x.GetString("namespace"), c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			args, helmArgs := parseExtraArgs(args)

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if err := sq.UpdateLocalDependencies(cmd.Context(), x.GetInt("parallel")); err != nil {
				return errors.Wrap(err, "failed to update dependencies")
			}

			out, err := sq.Template(cmd.Context(), helmArgs, x.GetInt("parallel"))
			if err != nil {
				return errors.Wrap(err, "failed to render template")
			}

			switch {
			case x.GetBool("raw"):
				pterm.Println(out)
			case x.GetString("output") != "":
				pterm.Info.Printfln("💾 | writing outut to %s", x.GetString("output"))
				return os.WriteFile(x.GetString("output"), []byte(out), 0600)
			default:
				pterm.Println(util.Highlight(out))
			}

			return nil
		},
	}

	flags := cmd.Flags()

	addParallelFlag(flags, x)
	addNamespaceFlag(flags, x)
	addOutputFlag(flags, x)
	addTagsFlag(flags, x)
	addRawFlag(flags, x)

	return cmd
}
