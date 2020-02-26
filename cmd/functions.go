package main

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"gopkg.in/yaml.v2"
)

type Docker struct {
	File    string
	Context string
	Options string
	Image   string
}

type Service struct {
	Name   string
	Docker Docker
}

type ServiceGroupConfig struct {
	Name     string
	Revision int
	Services []Service
}

func parseConfig(fileName string) (ServiceGroupConfig, error) {
	sgc := ServiceGroupConfig{}
	configFile, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Printf("Error reading config file: %s\n", err)
	} else {
		err = yaml.Unmarshal(configFile, &sgc)
		if err != nil {
			fmt.Printf("Error parsing config file: %s\n", err)
		}
	}
	if err == nil {
		fmt.Println("Successfully parsed config")
	}
	return sgc, err
}

func buildImage(service Service) (string, error) {
	fmt.Println("Building image for service " + service.Name)
	cmd := exec.Command("docker", "build", "-f", service.Docker.File, service.Docker.Context, "--no-cache")
	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)
	fmt.Println("\t" + output)

	image := ""
	if err == nil {
		image = output[strings.LastIndex(output, " ")+1 : len(output)-2]
		//there is \n\t at the end
	}
	return image, err
}

func Add(configFileName string) {
	//parse services configuration
	config, configErr := parseConfig(configFileName)
	if configErr != nil {
		return
	}

	//build docker images
	for i, service := range config.Services {
		image, buildErr := buildImage(service)
		if buildErr != nil {
			return
		}
		config.Services[i].Docker.Image = image
	}

	//generate charts
	//deploy with helm
}
