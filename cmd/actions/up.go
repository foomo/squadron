package actions

import (
	"context"
	"fmt"
	"os/user"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/util"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	upCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes units to the registry")
	upCmd.Flags().BoolVar(&flagDiff, "diff", false, "preview upgrade as a coloured diff")
	upCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	upCmd.Flags().StringVar(&flagBuildArgs, "build-args", "", "additional docker buildx build args")
	upCmd.Flags().StringVar(&flagPushArgs, "push-args", "", "additional docker push args")
}

var upCmd = &cobra.Command{
	Use:     "up [UNIT...]",
	Short:   "installs the squadron or given units",
	Example: "  squadron up frontend backend --namespace demo --build --push -- --dry-run",
	RunE: func(cmd *cobra.Command, args []string) error {
		return up(cmd.Context(), args, cwd, flagNamespace, flagBuild, flagPush, flagDiff, flagParallel, flagFiles)
	},
}

func up(ctx context.Context, args []string, cwd, namespace string, build, push, diff bool, parallel int, files []string) error {
	sq := squadron.New(cwd, namespace, files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

	args, helmArgs := parseExtraArgs(args)

	unitsNames, err := parseUnitNames(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	if unitsNames != nil {
		if err := sq.FilterConfig(unitsNames); err != nil {
			return err
		}
	}

	if err := sq.RenderConfig(ctx); err != nil {
		return err
	}

	units, err := parseUnitArgs(args, sq.GetConfig().Units)
	if err != nil {
		return err
	}

	if build {
		sem := semaphore.NewWeighted(int64(parallel))
		wg, wgCtx := errgroup.WithContext(ctx)

		_ = sq.GetConfig().Units.Iterate(func(n string, u *squadron.Unit) error {
			name := n
			unit := u
			wg.Go(func() error {
				if err := sem.Acquire(wgCtx, 1); err != nil {
					return err
				}
				defer sem.Release(1)
				if out, err := unit.Build(wgCtx, sq.Name(), name, strings.Split(flagBuildArgs, " ")); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})
		if err := wg.Wait(); err != nil {
			return err
		}
	}

	if push {
		sem := semaphore.NewWeighted(int64(parallel))
		wg, wgCtx := errgroup.WithContext(ctx)

		_ = sq.GetConfig().Units.Iterate(func(n string, u *squadron.Unit) error {
			name := n
			unit := u
			wg.Go(func() error {
				if err := sem.Acquire(wgCtx, 1); err != nil {
					return err
				}
				defer sem.Release(1)
				if out, err := unit.Push(wgCtx, sq.Name(), name, strings.Split(flagPushArgs, " ")); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})

		if err := wg.Wait(); err != nil {
			return err
		}
	}

	if err := sq.Generate(ctx, units); err != nil {
		return err
	}

	username := "unknown"
	if value, err := util.NewCommand("git").Args("config", "user.name").Run(ctx); err == nil {
		username = strings.TrimSpace(value)
	} else if value, err := user.Current(); err == nil {
		username = strings.TrimSpace(value.Name)
	}

	commit := ""
	if value, err := util.NewCommand("git").Args("rev-parse", "--short", "HEAD").Run(ctx); err == nil {
		commit = strings.TrimSpace(value)
	}

	if !diff {
		return sq.Up(ctx, units, helmArgs, username, version, commit)
	} else if out, err := sq.Diff(ctx, units, helmArgs); err != nil {
		return err
	} else {
		fmt.Println(out)
	}

	return nil
}
