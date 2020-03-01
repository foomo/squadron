package configurd

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type Service struct {
	Name   string `yaml:"name"`
	Docker Docker `yaml:"docker"`
	Chart  string `yaml:"chart"`
}

type Docker struct {
	File    string `yaml:"file"`
	Context string `yaml:"context"`
	Options string `yaml:"options"`
	Image   string `yaml:"image"`
}

func LoadService(reader io.Reader) (Service, error) {
	var wrapper struct {
		Service Service `yaml:"service"`
	}
	if err := yaml.NewDecoder(reader).Decode(&wrapper); err != nil {
		return Service{}, fmt.Errorf("could not decode service: %w", err)
	}
	return wrapper.Service, nil
}

func (s Service) Build(ctx context.Context, tag string) (string, error) {
	args := []string{
		"build",
		"-f", s.Docker.File,
		"-t", fmt.Sprintf("%s:%s", s.Docker.Image, tag),
		".",
	}

	cmd := exec.Command("docker", args...)
	cmd.Dir = s.Docker.Context

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)
	if err != nil {
		return output, err
	}
	return output, nil
}
