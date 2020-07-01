package util

import (
	"github.com/sirupsen/logrus"
)

type GoCmd struct {
	*Cmd
}

func NewGoCommand(l *logrus.Entry) *GoCmd {
	return &GoCmd{NewCommand(l, "go")}
}

func (c GoCmd) Build(workDir, output, input string, flags ...string) *Cmd {
	return c.Args("build", "-o", output).Args(flags...).Args(input)
}
