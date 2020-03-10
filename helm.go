package configurd

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	"gopkg.in/yaml.v2"
)

type ServiceGroupItem struct {
	ServiceName  string `yaml:"service"`
	Overrides    interface{}
	namespace    string
	serviceGroup string
	chart        string
}

func generateYaml(log Logger, path string, data interface{}) error {
	log.Printf("Generating yaml file: %v", path)
	out, marshalErr := yaml.Marshal(data)
	if marshalErr != nil {
		return marshalErr
	}
	file, crateErr := os.Create(path)
	if crateErr != nil {
		return crateErr
	}
	defer file.Close()
	_, writeErr := file.Write(out)
	if writeErr != nil {
		return writeErr
	}
	return nil
}

func generate(log Logger, sgi ServiceGroupItem, outputDir, chartPath string) error {
	dir := path.Join(outputDir, sgi.ServiceName)
	log.Printf("Creating dir: %v", dir)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return fmt.Errorf("could not create output dir: %w", err)
	}

	log.Printf("Copying %v to %v", chartPath, dir)
	err := filepath.Walk(chartPath, func(source string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chartPath == source {
			return nil
		}

		destination := path.Join(dir, info.Name())
		if err := copy.Copy(source, destination); err != nil {
			return err
		}
		if info.IsDir() {
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("could not copy template files: %w", err)
	}

	err = generateYaml(log, path.Join(dir, defaultOverridesFile), sgi.Overrides)
	if err != nil {
		return fmt.Errorf("could not generate %v: %w", defaultOverridesFile, err)
	}
	return nil
}

func helmInstall(log Logger, s Service, namespace, outputDir, tag string) (string, error) {
	log.Printf("Running helm install for service: %v", s.Name)
	chartPath := path.Join(outputDir, s.Name)
	cmdArgs := []string{
		"install", s.Name, chartPath,
		"-f", path.Join(chartPath, defaultOverridesFile),
		"-n", namespace,
		"--set", fmt.Sprintf("nameOverride=%v", s.Name),
		"--set", fmt.Sprintf("fullnameOverride=%v", s.Name),
		"--set", fmt.Sprintf("image.repository=%v", s.Build.Image),
		"--set", fmt.Sprintf("image.tag=%v", tag),
	}
	cmd := exec.Command("helm", cmdArgs...)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)

	if err != nil {
		return "", fmt.Errorf("could not install a helm chart for service %v output: \n%v", s.Name, output)
	}
	return output, nil
}

func helmUninstall(log Logger, sgi ServiceGroupItem) (string, error) {
	log.Printf("Running helm uninstall for service: %v", sgi.ServiceName)
	cmdArgs := []string{
		"uninstall", "-n", sgi.namespace, sgi.ServiceName,
	}
	cmd := exec.Command("helm", cmdArgs...)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)

	if err != nil {
		return "", fmt.Errorf("could not uninstall a helm chart for namespace: %v, service group: %v, output: \n%v", sgi.namespace, sgi.serviceGroup, output)
	}
	return output, nil
}

func (c Configurd) Install(log Logger, sgis []ServiceGroupItem, deploymentDir, tag string) (string, error) {
	outputDir := path.Join(defaultOutputDir, deploymentDir)

	log.Printf("Removing dir: %v", outputDir)
	if err := os.RemoveAll(outputDir); err != nil {
		return "", fmt.Errorf("could not clean workdir directory: %w", err)
	}

	log.Printf("Creating dir: %v", outputDir)
	if err := os.MkdirAll(outputDir, 0744); err != nil {
		return "", fmt.Errorf("could not create a workdir directory: %w", err)
	}
	for _, sgi := range sgis {
		err := generate(log, sgi, outputDir, path.Join(defaultChartDir, sgi.chart))
		if err != nil {
			return "", err
		}
	}

	var output []string
	for _, sgi := range sgis {
		s, err := c.Service(sgi.ServiceName)
		if err != nil {
			return "", err
		}
		out, err := helmInstall(log, s, sgi.namespace, outputDir, tag)
		if err != nil {
			return "", err
		}
		output = append(output, out)
	}

	return strings.Join(output, "\n"), nil
}

func (c Configurd) Uninstall(log Logger, sgis []ServiceGroupItem) (string, error) {
	var output []string
	for _, sgi := range sgis {
		out, err := helmUninstall(log, sgi)
		if err != nil {
			return "", err
		}
		output = append(output, out)
	}
	return strings.Join(output, "\n"), nil
}
