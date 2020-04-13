package configurd

import (
	"fmt"
	"strings"
)

type Service struct {
	Name  string `yaml:"-"`
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
	Build string `yaml:"build"`
	Chart string `yaml:"chart"`
}

func (s Service) RunBuild(log Logger, dir, tag string, verbose bool) (string, error) {
	if s.Build == "" {
		return "", ErrBuildNotConfigured
	}
	args := strings.Split(s.Build, " ")
	if args[0] == "docker" {
		args = append(strings.Split(s.Build, " "), "-t", fmt.Sprintf("%v:%v", s.Image, s.Tag))
	}
	log.Printf("Building service: %v", s.Name)

	output, err := runCommand(dir, args...)
	if err != nil {
		return output, err
	}
	logOutput(log, verbose, output)
	return output, nil
}
