package actions

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/foomo/squadron"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes built squadron units to the registry")
	buildCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
}

var buildCmd = &cobra.Command{
	Use:     "build [UNIT...]",
	Short:   "build or rebuild squadron units",
	Example: "  squadron build frontend backend",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {
		return build(cmd.Context(), args, cwd, flagFiles, flagPush, flagParallel)
	},
}

func build(ctx context.Context, args []string, cwd string, files []string, push bool, parallel int) error {
	sq := squadron.New(cwd, "", files)

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
				if out, err := unit.Build(wgCtx, sq.Name(), name); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
		}

		if err := wg.Wait(); err != nil {
			return err
		}
	}

	if push {
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
				if out, err := unit.Push(wgCtx, sq.Name(), name); err != nil {
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
