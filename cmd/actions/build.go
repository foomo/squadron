package actions

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/foomo/squadron"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "pushes built squadron units to the registry")
	buildCmd.Flags().IntVar(&flagParallel, "parallel", 1, "run command in parallel")
	buildCmd.Flags().StringVar(&flagBuildArgs, "build-args", "", "additional docker buildx build args")
	buildCmd.Flags().StringVar(&flagPushArgs, "push-args", "", "additional docker push args")
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
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(parallel)

		_ = squadron.Units(units).Iterate(func(n string, u *squadron.Unit) error {
			name := n
			unit := u
			g.Go(func() error {
				if out, err := unit.Build(gctx, sq.Name(), name, strings.Split(flagBuildArgs, " ")); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})
		err := g.Wait()
		if err != nil {
			return err
		}
	}

	if push {
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(parallel)

		_ = squadron.Units(units).Iterate(func(n string, u *squadron.Unit) error {
			name := n
			unit := u
			g.Go(func() error {
				if out, err := unit.Push(gctx, sq.Name(), name, strings.Split(flagPushArgs, " ")); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})

		if err := g.Wait(); err != nil {
			return err
		}
	}

	return nil
}
