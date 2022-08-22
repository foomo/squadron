package runner

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestRunner(t *testing.T) {
	t.Run("run", func(t *testing.T) {
		runner := Runner{}
		completed := atomic.NewInt32(0)
		count := 100
		for i := 0; i < count; i++ {
			runner.Add(func(ctx context.Context) error {
				completed.Inc()
				return nil
			})
		}
		err := runner.Run(context.Background(), 4)
		require.NoError(t, err)
		require.Equal(t, count, int(completed.Load()))
	})
	t.Run("error", func(t *testing.T) {
		runner := Runner{}
		completed := atomic.NewInt32(0)

		success := func(ctx context.Context) error {
			completed.Inc()
			return nil
		}
		fail := func(ctx context.Context) error {
			return errors.New("fail")
		}

		runner.Add(success)
		runner.Add(success)
		runner.Add(fail)
		runner.Add(success)

		err := runner.Run(context.Background(), 1)
		require.Error(t, err)
		require.Equal(t, 2, int(completed.Load()))
	})
}
