package configurd

import (
	"fmt"
	"os/exec"
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

func (s Service) RunBuild(log Logger, dir, tag string) (string, error) {
	args := append(strings.Split(s.Build.Command, " "), "-t", fmt.Sprintf("%v:%v", s.Build.Image, tag))
	log.Printf("Running command: %v", strings.Join(args, " "))

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)
	if err != nil {
		return output, err
	}
	return output, nil
}
