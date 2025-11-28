package pterm

import (
	"strings"
	"time"

	"github.com/pterm/pterm"
)

type NoopSpinner struct {
	prefix  string
	stopped bool
	start   time.Time
	log     []string
}

func NewNoopSpinner(prefix string) *NoopSpinner {
	return &NoopSpinner{
		prefix: prefix,
	}
}

func (s *NoopSpinner) Start(message ...string) {
	pterm.Info.Println(s.message(message...))
}

func (s *NoopSpinner) Play() {
	s.start = time.Now()
}

func (s *NoopSpinner) Info(message ...string) {
	s.stopped = true
	pterm.Info.Println(s.message(message...))
}

func (s *NoopSpinner) Warning(message ...string) {
	s.stopped = true
	pterm.Warning.Println(s.message(message...))
}

func (s *NoopSpinner) Fail(message ...string) {
	s.stopped = true
	pterm.Error.Println(s.message(message...))
}

func (s *NoopSpinner) Success(message ...string) {
	s.stopped = true
	pterm.Success.Println(s.message(message...))
}

func (s *NoopSpinner) Write(p []byte) (int, error) {
	var lines []string

	for _, line := range strings.Split(string(p), "\n") {
		if line := strings.TrimSpace(line); len(line) > 0 {
			lines = append(lines, line)
		}
	}

	s.log = append(s.log, lines...)
	// pterm.UpdateText.Println(s.message())
	return len(p), nil
}

func (s *NoopSpinner) message(message ...string) string {
	msg := []string{s.prefix}
	if !s.start.IsZero() && s.stopped {
		msg[0] += " ⏱ " + time.Since(s.start).Round(time.Millisecond).String()
	}

	if value := strings.Join(message, " "); len(value) > 0 {
		msg = append(msg, value)
	}

	if pterm.PrintDebugMessages {
		msg = append(msg, s.log...)
	}

	return strings.Join(msg, "\n  ")
}
