package util

import (
	"fmt"
	"io"
)

type DockerCmd struct {
	Cmd
	Options []string
}

func NewDockerCommand() *DockerCmd {
	return &DockerCmd{*NewCommand("docker"), []string{}}
}

func (c *DockerCmd) Bake(in io.Reader) *Cmd {
	return c.Stdin(in).Args("buildx", "bake", "--allow", "fs.read=*", "all", "-f", "-")
}

func (c *DockerCmd) Build(workDir string) *Cmd {
	return c.Cwd(workDir).Args("buildx", "build", ".")
}

func (c *DockerCmd) Push(image, tag string) *Cmd {
	return c.Args("push", fmt.Sprintf("%s:%s", image, tag))
}
