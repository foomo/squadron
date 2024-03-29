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

func (c *DockerCmd) Build(workDir string) *Cmd {
	return c.Cwd(workDir).Args("buildx", "build", ".")
}

func (c *DockerCmd) Push(image, tag string) *Cmd {
	return c.Args("push", fmt.Sprintf("%s:%s", image, tag))
}
