package util

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type DockerCmd struct {
	Cmd
}

func NewDockerCommand(l *logrus.Entry) *DockerCmd {
	return &DockerCmd{*NewCommand(l, "docker")}
}

func (c DockerCmd) Build(workDir string, options ...string) *Cmd {
	return c.Args("build", workDir).Args(options...)
}

func (c DockerCmd) Push(image, tag string, options ...string) (string, error) {
	return c.Args("push", fmt.Sprintf("%v:%v", image, tag)).Args(options...).Run()
}
