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

	wg, wgCtx := errgroup.WithContext(ctx)
	for i := 0; i < routines; i++ {
		wg.Go(func() error {
			for task := range tasks {
				if wgCtx.Err() != nil {
					return wgCtx.Err()
				}
				if err := task(wgCtx); err != nil {
					return err
				}
			}
			return nil
		})
	}

	return wg.Wait()
}
