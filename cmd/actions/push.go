package actions

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/foomo/squadron"
)

func init() {
	pushCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "specifies the namespace")
	pushCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "builds or rebuilds units")
	pushCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	pushCmd.Flags().StringVar(&flagBuildArgs, "build-args", "", "additional docker buildx build args")
	pushCmd.Flags().StringVar(&flagPushArgs, "push-args", "", "additional docker push args")
}

var pushCmd = &cobra.Command{
	Use:     "push [UNIT...]",
	Short:   "pushes the squadron or given units",
	Example: "  squadron push frontend backend --namespace demo --build",
	RunE: func(cmd *cobra.Command, args []string) error {
		return push(cmd.Context(), args, cwd, flagNamespace, flagBuild, flagParallel, flagFiles)
	},
}

func push(ctx context.Context, args []string, cwd, namespace string, build bool, parallel int, files []string) error {
	sq := squadron.New(cwd, namespace, files)

	if err := sq.MergeConfigFiles(); err != nil {
		return err
	}

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

		for n, u := range units {
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
		}

		if err := wg.Wait(); err != nil {
			return err
		}
	}

	{
		sem := semaphore.NewWeighted(int64(parallel))
		wg, wgCtx := errgroup.WithContext(ctx)

		for n, u := range units {
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
		}

		if err := wg.Wait(); err != nil {
			return err
		}
	}

	return nil
}
