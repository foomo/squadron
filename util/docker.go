package util

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func NewDockerCommand(l *logrus.Entry) (*CliCommand, error) {
	return NewCliCommand(l, "docker")
}

func (cc CliCommand) Build(workDir string, options ...string) *Cmd {
	cmd := append(cc.GetCommand(), "build", workDir)
	cmd = append(cmd, options...)
	return Command(cc.l, cmd...)
}

func (cc CliCommand) Push(name, tag string, options ...string) (string, error) {
	cmd := append(cc.GetCommand(), "push", fmt.Sprintf("%v:%v", name, tag))
	cmd = append(cmd, options...)
	return Command(cc.l, cmd...).Run()
}
