package configurd

import (
	"fmt"
	"strings"
)

type Build struct {
	Command string `yaml:"command"`
	Image   string `yaml:"image"`
}

type Service struct {
	Name  string
	Build Build  `yaml:"build"`
	Chart string `yaml:"chart"`
}

func (s Service) RunBuild(log Logger, dir, tag string, verbose bool) (string, error) {
	args := strings.Split(s.Build.Command, " ")
	if args[0] == "docker" {
		args = append(strings.Split(s.Build.Command, " "), "-t", fmt.Sprintf("%v:%v", s.Build.Image, tag))
	}
	log.Printf("Building service: %v", s.Name)

	output, err := runCommand(dir, args...)

	if err != nil {
		return output, err
	}
	logOutput(log, verbose, output)
	return output, nil
}
