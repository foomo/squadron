package util

import (
	"context"
	"fmt"
	"os"
)

type DockerCmd struct {
	Cmd
	Options []string
}

func NewDockerCommand() *DockerCmd {
	return &DockerCmd{*NewCommand("docker"), []string{}}
}

func (c *DockerCmd) Build(workDir string) *Cmd {
	args := []string{"build"}
	if platform := os.Getenv("SQUADRON_DOCKER_BUILDX"); platform != "" {
		args = []string{"buildx", "build", "--platform", platform}
	}
	args = append(args, ".")
	return c.Cwd(workDir).Args(args...)
}

func (c *DockerCmd) Push(ctx context.Context, image, tag string) (string, error) {
	return c.Args("push", fmt.Sprintf("%s:%s", image, tag)).Run(ctx)
}
