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

func NewSchema(c *viper.Viper) *cobra.Command {
	x := viper.New()
	cmd := &cobra.Command{
		Use:     "schema [SQUADRON]",
		Short:   "generate squadron json schema",
		Example: "  squadron schema",
		Args:    cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, "", c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			out, err := sq.RenderSchema(cmd.Context(), x.GetString("base-schema"))
			if err != nil {
				return errors.Wrap(err, "failed to render schema")
			}

			if output := x.GetString("output"); output != "" {
				pterm.Info.Printfln("💾 | writing outut to %s", output)

				if err := os.WriteFile(output, []byte(out), 0600); err != nil {
					return errors.Wrap(err, "failed to write schema")
				}

				return nil
			}

			if !x.GetBool("raw") {
				out = util.Highlight(out)
			}

			pterm.Println(out)

			return nil
		},
	}

	flags := cmd.Flags()

	addOutputFlag(flags, x)

	flags.String("base-schema", "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json", "Base schema to use")
	_ = x.BindPFlag("base-schema", flags.Lookup("base-schema"))

	addTagsFlag(flags, x)
	addRawFlag(flags, x)

	return cmd
}
