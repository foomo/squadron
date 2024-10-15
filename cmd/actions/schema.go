package actions

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	flagOutput     string
	flagBaseSchema string
)

func init() {
	schemaCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
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

		return sq.WriteSchema(cmd.Context(), flagBaseSchema, flagParallel)
	},
}
