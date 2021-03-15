package util

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type DockerCmd struct {
	Cmd
	Options []string
}

func NewDockerCommand(l *logrus.Entry) *DockerCmd {
	return &DockerCmd{*NewCommand(l, "docker"), []string{}}
}

func (c DockerCmd) Build(workDir string) (string, error) {
	return c.Args("build", workDir).Args(c.Options...).Run()
}

func (c *DockerCmd) Option(name, v string) *Cmd {
	if v == "" {
		return &c.Cmd
	}
	c.Options = append(c.Options, name, v)
	return &c.Cmd
}

func (c *DockerCmd) ListOption(name string, v []string) *Cmd {
	for _, i := range v {
		c.Options = append(c.Options, name, i)
	}
	return &c.Cmd
}

func (c DockerCmd) Push(image, tag string, options ...string) (string, error) {
	return c.Args("push", fmt.Sprintf("%v:%v", image, tag)).Args(options...).Run()
}
