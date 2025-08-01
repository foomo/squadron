package actions

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
				pterm.Info.Printfln("Writing JSON schema to %s", output)
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
	flags.String("output", "", "Output file")
	_ = x.BindPFlag("output", flags.Lookup("output"))

	flags.String("base-schema", "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json", "Base schema to use")
	_ = x.BindPFlag("base-schema", flags.Lookup("base-schema"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = x.BindPFlag("tags", flags.Lookup("tags"))

	flags.Bool("raw", false, "print raw output without highlighting")
	_ = x.BindPFlag("raw", flags.Lookup("raw"))

	return cmd
}
