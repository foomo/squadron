package actions

import (
	"os/user"
	"strings"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/internal/util"
	"github.com/spf13/cobra"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	upCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes units to the registry")
	upCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	upCmd.Flags().StringSliceVar(&flagBuildArgs, "build-args", nil, "additional docker buildx build args")
	upCmd.Flags().StringSliceVar(&flagPushArgs, "push-args", nil, "additional docker push args")
}

var upCmd = &cobra.Command{
	Use:     "up [SQUADRON] [UNIT...]",
	Short:   "installs the squadron or given units",
	Example: "  squadron up storefinder frontend backend --namespace demo --build --push -- --dry-run",
	RunE: func(cmd *cobra.Command, args []string) error {
		sq := squadron.New(cwd, flagNamespace, flagFiles)

		if err := sq.MergeConfigFiles(); err != nil {
			return err
		}

		args, helmArgs := parseExtraArgs(args)

		squadronName, unitNames := parseSquadronAndUnitNames(args)
		if err := sq.FilterConfig(squadronName, unitNames); err != nil {
			return err
		}

		if err := sq.RenderConfig(cmd.Context()); err != nil {
			return err
		}

		if flagBuild {
			if err := sq.Build(cmd.Context(), flagBuildArgs, flagParallel); err != nil {
				return err
			}
		}

		if flagPush {
			if err := sq.Push(cmd.Context(), flagPushArgs, flagParallel); err != nil {
				return err
			}
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
