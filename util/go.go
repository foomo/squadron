package util

import (
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

type GoCmd struct {
	Cmd
}

func NewGoCommand(l *logrus.Entry) *GoCmd {
	return &GoCmd{*NewCommand(l, "go")}
}

func (c GoCmd) Build(workDir, output, input string, flags ...string) *Cmd {
	relInput := strings.TrimPrefix(input, workDir+string(filepath.Separator))
	return c.Args("build", "-o", output).Cwd(workDir).Args(flags...).Args(relInput)
}
