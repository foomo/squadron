package pterm

import (
	"io"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type StandardSpinner struct {
	printer *pterm.SpinnerPrinter
	prefix  string
	stopped bool
	start   time.Time
	log     []string
}

func NewStandardSpinner(writer io.Writer, prefix string) *StandardSpinner {
	return &StandardSpinner{
		printer: pterm.DefaultSpinner.WithWriter(writer).
			WithDelay(500*time.Millisecond).
			WithSequence("▀  ", " ▀ ", " ▄ ", "▄  ").
			WithShowTimer(false),
		prefix: prefix,
	}
}

func (s *StandardSpinner) Start(message ...string) {
	var err error
	if s.printer, err = s.printer.Start(s.message(message...)); err != nil {
		pterm.Fatal.Println(err)
	}
	s.start = time.Now()
}

func (s *StandardSpinner) Info(message ...string) {
	s.stopped = true
	s.printer.Info(s.message(message...))
}

func (s *StandardSpinner) Warning(message ...string) {
	s.stopped = true
	s.printer.Warning(s.message(message...))
}

func (s *StandardSpinner) Fail(message ...string) {
	s.stopped = true
	s.printer.Fail(s.message(message...))
}

func (s *StandardSpinner) Success(message ...string) {
	s.stopped = true
	s.printer.Success(s.message(message...))
}

func (s *StandardSpinner) Write(p []byte) (int, error) {
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

func (s *StandardSpinner) message(message ...string) string {
	msg := []string{s.prefix}
	if !s.start.IsZero() && s.stopped {
		msg[0] += " ⏱ " + time.Since(s.start).Round(0).String()
	}
	if value := strings.Join(message, " "); len(value) > 0 {
		msg = append(msg, value)
	}
	if pterm.PrintDebugMessages {
		msg = append(msg, s.log...)
	}
	m := pterm.GetTerminalWidth() - 10
	for i, line := range msg {
		if len(line) > m {
			msg[i] = line[:m] + "…"
		}
	}
	return strings.Join(msg, "\n  ")
}
