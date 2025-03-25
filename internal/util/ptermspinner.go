package util

import (
	"context"
	"io"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type contextKey string

const contextKeyPTermSpinner contextKey = "PtermSpinner"

type PTermSpinner struct {
	printer *pterm.SpinnerPrinter
	prefix  string
	stopped bool
	start   time.Time
	log     []string
}

func NewPTermSpinner(writer io.Writer, prefix string) *PTermSpinner {
	return &PTermSpinner{
		printer: pterm.DefaultSpinner.WithWriter(writer).
			WithDelay(500*time.Millisecond).
			WithSequence("▀  ", " ▀ ", " ▄ ", "▄  ").
			WithShowTimer(false),
		prefix: prefix,
	}
}

func PTermSpinnerFromContext(ctx context.Context) *PTermSpinner {
	if value, ok := ctx.Value(contextKeyPTermSpinner).(*PTermSpinner); ok {
		return value
	}
	return nil
}

func (s *PTermSpinner) Inject(ctx context.Context) context.Context {
	return context.WithValue(ctx, contextKeyPTermSpinner, s)
}

func (s *PTermSpinner) Start(message ...string) {
	var err error
	if s.printer, err = s.printer.Start(s.message(message...)); err != nil {
		pterm.Fatal.Println(err)
	}
	s.start = time.Now()
}

func (s *PTermSpinner) Info(message ...string) {
	s.stopped = true
	s.printer.Info(s.message(message...))
}

func (s *PTermSpinner) Warning(message ...string) {
	s.stopped = true
	s.printer.Warning(s.message(message...))
}

func (s *PTermSpinner) Fail(message ...string) {
	s.stopped = true
	s.printer.Fail(s.message(message...))
}

func (s *PTermSpinner) Success(message ...string) {
	s.stopped = true
	s.printer.Success(s.message(message...))
}

func (s *PTermSpinner) Write(p []byte) (int, error) {
	var lines []string
	for _, line := range strings.Split(string(p), "\n") {
		if line := strings.TrimSpace(line); len(line) > 0 {
			lines = append(lines, line)
		}
	}
	s.log = append(s.log, lines...)
	// s.printer.UpdateText(s.message())
	return len(p), nil
}

func (s *PTermSpinner) message(message ...string) string {
	msg := s.prefix
	if !s.start.IsZero() && s.stopped {
		msg += " ⏱ " + time.Since(s.start).Round(0).String()
	}
	if value := strings.Join(message, " "); len(value) > 0 {
		msg += "\n" + value
	}
	if pterm.PrintDebugMessages {
		msg += "\n" + strings.Join(s.log, "\n")
	}
	return strings.TrimSpace(strings.ReplaceAll(msg, "\n", "\n "))
}
