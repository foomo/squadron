package runner

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type Task func(ctx context.Context) error

type Runner struct {
	tasks []Task
}

func (r *Runner) Add(task Task) {
	r.tasks = append(r.tasks, task)
}

func (r *Runner) Run(ctx context.Context, routines int) error {
	tasks := make(chan Task, len(r.tasks))
	go func() {
		defer close(tasks)
		for _, t := range r.tasks {
			tasks <- t
		}
	}()

	g, gctx := errgroup.WithContext(ctx)
	for i := 0; i < routines; i++ {
		g.Go(func() error {
			for task := range tasks {
				if gctx.Err() != nil {
					return gctx.Err()
				}
				if err := task(gctx); err != nil {
					return err
				}
			}
			return nil
		})
	}

	return g.Wait()
}
