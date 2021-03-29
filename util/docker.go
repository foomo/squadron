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

func (c *DockerCmd) Build(workDir string) (string, error) {
	c.l.Infof("Running docker build for %q", workDir)
	return c.Args("build", workDir).Args(c.Options...).Run()
}

func (c *DockerCmd) Push(image, tag string, options ...string) (string, error) {
	c.l.Infof("Running docker push for %s:%s", image, tag)
	return c.Args("push", fmt.Sprintf("%s:%s", image, tag)).Args(options...).Run()
}

func (c *DockerCmd) ImageExists(image, tag string) (bool, error) {
	c.l.Infof("Checking image exists for %s:%s", image, tag)
	ret, err := c.Args("images", "--filter", fmt.Sprintf("reference='%s:%s'", image, tag)).Run()
	return ret != "", err
}
