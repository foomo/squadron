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

type ServiceItem struct {
	ServiceName string `yaml:"service"`
	Overrides   interface{}
	namespace   string
	group       string
	chart       string
}

type JobItem struct {
	ServiceName string `yaml:"service"`
	Overrides   interface{}
	namespace   string
	group       string
	chart       string
}

func generateYaml(log Logger, path string, data interface{}) error {
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

func generate(log Logger, si ServiceItem, basePath, outputDir, chart string) error {
	outputPath := path.Join(basePath, defaultOutputDir, outputDir, si.ServiceName)
	chartPath := path.Join(basePath, defaultChartDir, chart)

	log.Printf("Entering dir: %v", path.Join(basePath, defaultOutputDir, outputDir))
	log.Printf("Creating dir: %q", si.ServiceName)
	if err := os.MkdirAll(outputPath, 0744); err != nil {
		return fmt.Errorf("could not create output dir: %w", err)
	}

	log.Printf("Copying chart: %v to dir: %q", chart, si.ServiceName)
	err := filepath.Walk(chartPath, func(source string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chartPath == source {
			return nil
		}

		destination := path.Join(outputPath, info.Name())
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

	log.Printf("Generating yaml file: %q", path.Join(si.ServiceName, defaultOverridesFile))
	err = generateYaml(log, path.Join(outputPath, defaultOverridesFile), si.Overrides)
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

func helmUninstall(log Logger, si ServiceItem) (string, error) {
	log.Printf("Running helm uninstall for service: %v", si.ServiceName)
	cmdArgs := []string{
		"uninstall", "-n", si.namespace, si.ServiceName,
	}
	cmd := exec.Command("helm", cmdArgs...)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)

	if err != nil {
		return "", fmt.Errorf("could not uninstall a helm chart for namespace: %v, service group: %v, output: \n%v", si.namespace, si.group, output)
	}
	return output, nil
}

func (c Configurd) Install(log Logger, sis []ServiceItem, basePath, outputDir, tag string, verbose bool) (string, error) {
	log.Printf("Installing services")
	wordkDir := path.Join(basePath, defaultOutputDir)
	outputPath := path.Join(basePath, defaultOutputDir, outputDir)

	log.Printf("Entering dir: %q", wordkDir)
	log.Printf("Removing dir: %q", outputDir)
	if err := os.RemoveAll(outputPath); err != nil {
		return "", fmt.Errorf("could not clean workdir directory: %w", err)
	}

	log.Printf("Creating dir: %q", outputDir)
	if err := os.MkdirAll(outputPath, 0744); err != nil {
		return "", fmt.Errorf("could not create a workdir directory: %w", err)
	}
	for _, si := range sis {
		err := generate(log, si, basePath, outputDir, si.chart)
		if err != nil {
			return "", err
		}
	}

	var output []string
	for _, si := range sis {
		s, err := c.Service(si.ServiceName)
		if err != nil {
			return "", err
		}
		out, err := helmInstall(log, s, si.namespace, outputPath, tag)
		if err != nil {
			return "", err
		}
		logOutput(log, verbose, out)
		output = append(output, out)
	}

	return strings.Join(output, "\n"), nil
}

func (c Configurd) Uninstall(log Logger, sis []ServiceItem, verbose bool) (string, error) {
	var outputs []string
	for _, si := range sis {
		out, err := helmUninstall(log, si)
		if err != nil {
			return "", err
		}
		outputs = append(outputs, out)
		logOutput(log, verbose, out)
	}
	return strings.Join(outputs, "\n"), nil
}
