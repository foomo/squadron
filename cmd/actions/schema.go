package actions

import (
	"fmt"
	"os"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	flagOutput     string
	flagBaseSchema string
)

func init() {
	schemaCmd.Flags().StringVar(&flagOutput, "output", "", "Output file")
	schemaCmd.Flags().StringVar(&flagBaseSchema, "base-schema", "https://raw.githubusercontent.com/foomo/squadron/refs/heads/main/squadron.schema.json", "Base schema to use")
	schemaCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var schemaCmd = &cobra.Command{
	Use:     "schema [SQUADRON]",
	Short:   "generate squadron schemas",
	Example: "  squadron schema",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, "", flagFiles)

		if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to merge config files")
		}

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, flagTags); err != nil {
			return errors.Wrap(err, "failed to filter config")
		}

		js, err := sq.RenderSchema(cmd.Context(), flagBaseSchema)
		if err != nil {
			return errors.Wrap(err, "failed to render schema")
		}

		if flagOutput != "" {
			if err := os.WriteFile(flagOutput, []byte(js), 0600); err != nil {
				return errors.Wrap(err, "failed to write schema")
			}
		}

		fmt.Print(util.Highlight(js))

		return nil
	},
}
