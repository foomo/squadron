package actions

import (
	"os"

	"github.com/foomo/squadron"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewUp(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "up [SQUADRON] [UNIT...]",
		Short:   "installs the squadron or given units",
		Example: "  squadron up storefinder frontend backend --namespace demo --build --push -- --dry-run",
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

			if err := sq.RenderConfig(cmd.Context()); err != nil {
				return errors.Wrap(err, "failed to render config")
			}

			if c.GetBool("build") {
				if err := sq.Build(cmd.Context(), c.GetStringSlice("build-args"), c.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to build units")
				}
			}

			if c.GetBool("push") {
				if err := sq.Push(cmd.Context(), c.GetStringSlice("push-args"), c.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to push units")
				}
			}

			if err := sq.UpdateLocalDependencies(cmd.Context(), c.GetInt("parallel")); err != nil {
				return err
			}

			status := squadron.Status{
				Squadron: version,
				User:     "unknown",
			}
			if wd, err := os.Getwd(); err == nil {
				if value := os.Getenv("GIT_DIR"); value != "" {
					wd = value
				}
				if repo, err := git.PlainOpen(wd); err == nil {
					if c, err := repo.Config(); err == nil {
						status.User = c.User.Name
					}
					if ref, err := repo.Head(); err == nil {
						status.Branch = ref.Name().Short()
						status.Commit = ref.Hash().String()
						if tags, err := repo.Tags(); err == nil {
							_ = tags.ForEach(func(r *plumbing.Reference) error {
								if r.Hash() == ref.Hash() {
									status.Branch = r.Name().Short()
									return errors.New("found tag")
								}
								return nil
							})
						}
					}
				}
			}

			return sq.Up(cmd.Context(), helmArgs, status, c.GetInt("parallel"))
		},
	}
	flags := cmd.Flags()

	flags.StringP("namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	_ = c.BindPFlag("namespace", flags.Lookup("namespace"))

	flags.BoolP("build", "b", false, "builds or rebuilds units")
	_ = c.BindPFlag("build", flags.Lookup("build"))

	flags.BoolP("push", "p", false, "pushes units to the registry")
	_ = c.BindPFlag("push", flags.Lookup("push"))

	flags.Int("parallel", 1, "run command in parallel")
	_ = c.BindPFlag("parallel", flags.Lookup("parallel"))

	flags.StringArray("build-args", nil, "additional docker buildx build args")
	_ = c.BindPFlag("build-args", flags.Lookup("build-args"))

	flags.StringArray("push-args", nil, "additional docker push args")
	_ = c.BindPFlag("push-args", flags.Lookup("push-args"))

	flags.StringSlice("tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
	_ = c.BindPFlag("tags", flags.Lookup("tags"))

	return cmd
}
