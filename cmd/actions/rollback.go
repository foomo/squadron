package actions

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/foomo/squadron"
)

func NewRollback(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "rollback [SQUADRON] [UNIT...]",
		Short:   "rolls back the squadron or given units",
		Example: "  squadron rollback storefinder frontend backend --namespace demo",
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, c.GetString("namespace"), c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			args, helmArgs := parseExtraArgs(args)

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, c.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			return sq.Rollback(cmd.Context(), c.GetString("revision"), helmArgs, c.GetInt("parallel"))
		},
	}

	flags := cmd.Flags()
	flags.Int("parallel", 1, "run command in parallel")
	_ = c.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.StringP("namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	_ = c.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.StringP("revision", "r", "", "specifies the revision to roll back to")
	_ = c.BindPFlag("revision", flags.Lookup("revision"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = c.BindPFlag("tags", flags.Lookup("namespace"))

	return cmd
}
