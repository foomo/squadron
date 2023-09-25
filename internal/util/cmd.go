package util

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

type Cmd struct {
	// cmd           *exec.Cmd
	command       []string
	cwd           string
	env           []string
	stdin         io.Reader
	stdoutWriters []io.Writer
	stderrWriters []io.Writer
	wait          bool
	timeout       time.Duration
	preStartFunc  func() error
	postStartFunc func() error
	postEndFunc   func() error
}

func NewCommand(name string) *Cmd {
	return &Cmd{
		command: []string{name},
		wait:    true,
		env:     os.Environ(),
	}
}

func (c *Cmd) Args(args ...string) *Cmd {
	for _, arg := range args {
		if arg == "" {
			continue
		}
		c.command = append(c.command, arg)
	}
	return c
}

func (c *Cmd) Arg(name, v string) *Cmd {
	if name == "" || v == "" {
		return c
	}
	c.command = append(c.command, name, v)
	return c
}

func (c *Cmd) BoolArg(name string, v bool) *Cmd {
	if name == "" || !v {
		return c
	}
	c.command = append(c.command, name)
	return c
}

func (c *Cmd) ListArg(name string, vs []string) *Cmd {
	if name == "" {
		return c
	}
	for _, v := range vs {
		if v == "" {
			continue
		}
		c.command = append(c.command, name, v)
	}
	return c
}

func (c *Cmd) Cwd(path string) *Cmd {
	c.cwd = path
	return c
}

func (c *Cmd) Env(env ...string) *Cmd {
	c.env = append(c.env, env...)
	return c
}

func (c *Cmd) Stdin(r io.Reader) *Cmd {
	c.stdin = r
	return c
}

func (c *Cmd) Stdout(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stdoutWriters = append(c.stdoutWriters, w)
	return c
}

func (c *Cmd) Stderr(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stderrWriters = append(c.stderrWriters, w)
	return c
}

func (c *Cmd) Timeout(t time.Duration) *Cmd {
	c.timeout = t
	return c
}

func (c *Cmd) NoWait() *Cmd {
	c.wait = false
	return c
}

func (c *Cmd) PreStart(f func() error) *Cmd {
	c.preStartFunc = f
	return c
}

func (c *Cmd) PostStart(f func() error) *Cmd {
	c.postStartFunc = f
	return c
}

func (c *Cmd) PostEnd(f func() error) *Cmd {
	c.postEndFunc = f
	return c
}

func (c *Cmd) Run(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, c.command[0], c.command[1:]...) //nolint:gosec
	cmd.Env = append(os.Environ(), c.env...)
	if c.cwd != "" {
		cmd.Dir = c.cwd
	}
	pterm.Debug.Printfln("executing %s", cmd.String())

	combinedBuf := new(bytes.Buffer)
	traceWriter := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	cmd.Stdout = io.MultiWriter(append(c.stdoutWriters, combinedBuf, traceWriter)...)
	cmd.Stderr = io.MultiWriter(append(c.stderrWriters, combinedBuf, traceWriter)...)

	if c.preStartFunc != nil {
		pterm.Debug.Println("executing pre start func")
		if err := c.preStartFunc(); err != nil {
			return combinedBuf.String(), err
		}
	}

	if err := cmd.Start(); err != nil {
		return combinedBuf.String(), err
	}

	if c.postStartFunc != nil {
		pterm.Debug.Println("executing post start func")
		if err := c.postStartFunc(); err != nil {
			return combinedBuf.String(), err
		}
	}

	if c.wait {
		if c.timeout != 0 {
			timer := time.AfterFunc(c.timeout, func() {
				_ = cmd.Process.Kill()
			})
			defer timer.Stop()
		}

		if err := cmd.Wait(); err != nil {
			if c.timeout == 0 {
				return combinedBuf.String(), err
			}
		}
		if c.postEndFunc != nil {
			if err := c.postEndFunc(); err != nil {
				return combinedBuf.String(), err
			}
		}
	}

	return combinedBuf.String(), nil
}
