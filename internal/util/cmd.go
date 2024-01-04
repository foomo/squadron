package util

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
)

type Cmd struct {
	command       []string
	cwd           string
	env           []string
	stdin         io.Reader
	stdoutWriters []io.Writer
	stderrWriters []io.Writer
}

func NewCommand(name string) *Cmd {
	return &Cmd{
		command: []string{name},
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

func (c *Cmd) Run(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, c.command[0], c.command[1:]...) //nolint:gosec
	cmd.Env = append(os.Environ(), c.env...)
	if c.cwd != "" {
		cmd.Dir = c.cwd
	}
	pterm.Debug.Printfln("executing %s", cmd.String())

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	traceWriter := logrus.StandardLogger().WriterLevel(logrus.TraceLevel)

	if c.stdin != nil {
		cmd.Stdin = c.stdin
	}
	cmd.Stdout = io.MultiWriter(append(c.stdoutWriters, &stdout, traceWriter)...)
	cmd.Stderr = io.MultiWriter(append(c.stderrWriters, &stderr, traceWriter)...)

	err := cmd.Run()
	return stdout.String() + stderr.String(), err
}
