package actions

import (
	"os/user"
	"strings"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "set the namespace name or template (default, squadron-{{.Squadron}}-{{.Unit}})")
	upCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes units to the registry")
	upCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	upCmd.Flags().StringArrayVar(&flagBuildArgs, "build-args", nil, "additional docker buildx build args")
	upCmd.Flags().StringArrayVar(&flagPushArgs, "push-args", nil, "additional docker push args")
	upCmd.Flags().StringSliceVar(&flagTags, "tags", nil, "list of tags to include or exclude (can specify multiple or separate values with commas: tag1,tag2,-tag3)")
}

var upCmd = &cobra.Command{
	Use:     "up [SQUADRON] [UNIT...]",
	Short:   "installs the squadron or given units",
	Example: "  squadron up storefinder frontend backend --namespace demo --build --push -- --dry-run",
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, flagNamespace, flagFiles)

		if err := sq.MergeConfigFiles(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to merge config files")
		}

		args, helmArgs := parseExtraArgs(args)

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(cmd.Context(), squadronName, unitNames, flagTags); err != nil {
			return errors.Wrap(err, "failed to filter config")
		}

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return errors.Wrap(err, "failed to render config")
		}

		if flagBuild {
			if err := sq.Build(cmd.Context(), flagBuildArgs, flagParallel); err != nil {
				return errors.Wrap(err, "failed to build units")
			}
		}

		if flagPush {
			if err := sq.Push(cmd.Context(), flagPushArgs, flagParallel); err != nil {
				return errors.Wrap(err, "failed to push units")
			}
		}

		if err := sq.UpdateLocalDependencies(cmd.Context(), flagParallel); err != nil {
			return err
		}

		username := "unknown"
		if value, err := util.NewCommand("git").Args("config", "user.name").Run(cmd.Context()); err == nil {
			username = strings.TrimSpace(value)
		} else if value, err := user.Current(); err == nil {
			username = strings.TrimSpace(value.Name)
		}

		branch := ""
		if value, err := util.NewCommand("sh").Args("-c", "git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD || git rev-parse --short HEAD").Run(cmd.Context()); err == nil {
			branch = strings.TrimSpace(value)
		}
		commit := ""
		if value, err := util.NewCommand("sh").Args("-c", "git rev-parse --short HEAD").Run(cmd.Context()); err == nil {
			commit = strings.TrimSpace(value)
		}

		return sq.Up(cmd.Context(), helmArgs, username, version, commit, branch, flagParallel)
	},
}
