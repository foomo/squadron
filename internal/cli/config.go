package cli

import (
	"os"
	"strings"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewConfig(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "config [SQUADRON] [UNIT...]",
		Short:   "generate and view the squadron config",
		Example: "  squadron config storefinder frontend backend",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			files := c.GetStringSlice("file")
			sq := squadron.New(cwd, "", files)

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			pterm.Debug.Println(strings.Join(append([]string{"provided files"}, files...), "\n└ "))

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if !x.GetBool("no-render") {
				if err := sq.RenderConfig(cmd.Context()); err != nil {
					return errors.Wrap(err, "failed to render config")
				}
			}

			out := sq.ConfigYAML()

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

	flags.Bool("no-render", false, "don't render the config template")
	_ = x.BindPFlag("no-render", flags.Lookup("no-render"))

	addTagsFlag(flags, x)
	addOutputFlag(flags, x)
	addRawFlag(flags, x)

	return cmd
}
