package cli

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
	x := viper.New()

	cmd := &cobra.Command{
		Use:     "up [SQUADRON] [UNIT...]",
		Short:   "installs the squadron or given units",
		Example: "  squadron up storefinder frontend backend --namespace demo --build --push -- --dry-run",
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

			if x.GetBool("push") {
				if err := sq.Push(cmd.Context(), x.GetStringSlice("push-args"), x.GetInt("parallel")); err != nil {
					return errors.Wrap(err, "failed to push units")
				}
			}

			if err := sq.UpdateLocalDependencies(cmd.Context(), x.GetInt("parallel")); err != nil {
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

			return sq.Up(cmd.Context(), helmArgs, status, x.GetInt("parallel"))
		},
	}
	flags := cmd.Flags()

	addNamespaceFlag(flags, x)
	addBakeFlag(flags, x)
	addBuildFlag(flags, x)
	addPushFlag(flags, x)
	addParallelFlag(flags, x)
	addBakeArgsFlag(flags, x)
	addBuildArgsFlag(flags, x)
	addPushArgsFlag(flags, x)
	addTagsFlag(flags, x)

	return cmd
}
