package actions

import (
	"github.com/foomo/squadron"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPush(c *viper.Viper) *cobra.Command {
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "push [SQUADRON] [UNIT...]",
		Short:   "pushes the squadron or given units",
		Example: "  squadron push storefinder frontend backend --namespace demo --build",
		RunE: func(cmd *cobra.Command, args []string) error {
			sq := squadron.New(cwd, x.GetString("namespace"), c.GetStringSlice("file"))

			if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to merge config files")
			}

			squadronName, unitNames := parseSquadronAndUnitNames(args)
			if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, x.GetStringSlice("tags")); err != nil {
				return errors.Wrap(err, "failed to filter config")
			}

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if x.GetBool("bake") {
				bakefile, err := sq.Bakefile(cmd.Context())
				if err != nil {
					return errors.Wrap(err, "failed to bake units")
				}
				if err := sq.Bake(cmd.Context(), bakefile, x.GetStringSlice("bake-args")); err != nil {
					return errors.Wrap(err, "failed to bake units")
				}
			}

			if x.GetBool("build") {
				if err := sq.Build(cmd.Context(), x.GetStringSlice("build-args"), x.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to build units")
				}
			}

			return sq.Push(cmd.Context(), x.GetStringSlice("push-args"), x.GetInt("parallel"))
		},
	}

	flags := cmd.Flags()
	flags.StringP("namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	_ = x.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.Bool("build", false, "builds or rebuilds units")
	_ = x.BindPFlag("build", flags.Lookup("build"))

	flags.Bool("bake", false, "bakes or rebakes units")
	_ = x.BindPFlag("bake", flags.Lookup("bake"))

	flags.Int("parallel", 1, "run command in parallel")
	_ = x.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.StringArray("bake-args", nil, "additional docker buildx bake args")
	_ = x.BindPFlag("bake-args", flags.Lookup("bake-args"))

	flags.StringArray("build-args", nil, "additional docker buildx build args")
	_ = x.BindPFlag("build-args", flags.Lookup("build-args"))

	flags.StringArray("push-args", nil, "additional docker push args")
	_ = x.BindPFlag("push-args", flags.Lookup("push-args"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = x.BindPFlag("tags", flags.Lookup("tags"))

	return cmd
}
