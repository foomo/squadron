package cli

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/foomo/squadron"
)

func NewDown(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "down [SQUADRON] [UNIT...]",
		Short:   "uninstalls the squadron or given units",
		Example: "  squadron down storefinder frontend backend --namespace demo",
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

			return sq.Down(cmd.Context(), helmArgs, x.GetInt("parallel"))
		},
	}

	flags := cmd.Flags()

	addParallelFlag(flags, x)
	addNamespaceFlag(flags, x)
	addTagsFlag(flags, x)

	return cmd
}
