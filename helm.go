package configurd

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"
	"gopkg.in/yaml.v2"
)

type ServiceDeployment struct {
	ServiceName string `yaml:"service"`
	Overrides   interface{}
	namespace   string
	deployment  string
	chart       string
}

func generateYaml(path string, data interface{}) error {
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

func generate(sd ServiceDeployment, outputDir, chartPath string) error {
	dir := path.Join(outputDir, sd.ServiceName)
	if err := os.MkdirAll(dir, 0744); err != nil {
		return fmt.Errorf("could not create output dir: %w", err)
	}

	err := filepath.Walk(chartPath, func(source string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if chartPath == source {
			return nil
		}
		if err := copy.Copy(source, path.Join(dir, info.Name())); err != nil {
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

	err = generateYaml(path.Join(dir, defaultOverridesFile), sd.Overrides)
	if err != nil {
		return fmt.Errorf("could not generate %v: %w", defaultOverridesFile, err)
	}
	return nil
}

func helmInstall(serviceName, namespace, basePath string) error {
	chartPath := path.Join(basePath, defaultDeploymentDir, serviceName)
	cmdArgs := []string{
		"install", serviceName, chartPath,
		"-f", path.Join(chartPath, defaultOverridesFile),
		"-n", namespace,
		"--set", "fullnameOverride=" + serviceName,
	}
	cmd := exec.Command("helm", cmdArgs...)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)

	if err != nil {
		return fmt.Errorf("could not install a helm chart for service %v output: \n%v", serviceName, output)
	}
	return nil
}

func (c Configurd) Deploy(sds []ServiceDeployment, basePath string) (string, error) {
	if err := os.RemoveAll(defaultDeploymentDir); err != nil {
		return "", fmt.Errorf("could not clean deployment directory: %w", err)
	}

	if err := os.MkdirAll(defaultDeploymentDir, 0744); err != nil {
		return "", fmt.Errorf("could not create a deployments directory: %w", err)
	}
	for _, sd := range sds {
		err := generate(sd, defaultDeploymentDir, path.Join(defaultChartDir, sd.chart))
		if err != nil {
			return "", err
		}
		err = helmInstall(sd.ServiceName, sd.namespace, basePath)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func (c Configurd) Undeploy(sds []ServiceDeployment) error {
	for _, sd := range sds {
		cmdArgs := []string{
			"uninstall", "-n", sd.namespace, sd.ServiceName,
		}
		cmd := exec.Command("helm", cmdArgs...)

		out, err := cmd.CombinedOutput()
		output := strings.Replace(string(out), "\n", "\n\t", -1)

		if err != nil {
			return fmt.Errorf("could not uninstall a helm chart for namespace: %v, deployment: %v, output: \n%v", sd.namespace, sd.deployment, output)
		}
		log.Print(string(out))
	}
	return nil
}
