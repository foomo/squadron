package pterm

import (
	"context"
)

type Spinner interface {
	Start(message ...string)
	Play()
	Info(message ...string)
	Warning(message ...string)
	Fail(message ...string)
	Success(message ...string)
	Write(p []byte) (int, error)
	message(message ...string) string
}

func ContextWithSpinner(ctx context.Context, s Spinner) context.Context {
	return context.WithValue(ctx, contextKeySpinner, s)
}

func SpinnerFromContext(ctx context.Context) (Spinner, bool) {
	value, ok := ctx.Value(contextKeySpinner).(Spinner)
	return value, ok
}
