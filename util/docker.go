package util

import (
	"fmt"
)

type DockerCmd struct {
	Cmd
	Options []string
}

func NewDockerCommand() *DockerCmd {
	return &DockerCmd{*NewCommand("docker"), []string{}}
}

func (c *DockerCmd) Build(workDir string) (string, error) {
	return c.Cwd(workDir).Args("build", ".").Args(c.Options...).Run()
}

func (c *DockerCmd) Push(image, tag string, options ...string) (string, error) {
	return c.Args("push", fmt.Sprintf("%s:%s", image, tag)).Args(options...).Run()
}

func (c *DockerCmd) ImageExists(image, tag string) (bool, error) {
	ret, err := c.Args("images", "--quiet", fmt.Sprintf("%s:%s", image, tag)).Run()
	return ret != "", err
}
