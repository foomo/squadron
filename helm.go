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
	chart       string
	namespace   string
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

func generate(sd ServiceDeployment, outputDir, chartPath, tag string) error {
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

func (c Configurd) Deploy(sds []ServiceDeployment, basePath, tag string) (string, error) {
	if err := os.RemoveAll(defaultDeploymentDir); err != nil {
		return "", fmt.Errorf("could not clean deployment directory: %w", err)
	}

	if err := os.MkdirAll(defaultDeploymentDir, 0744); err != nil {
		return "", fmt.Errorf("could not create a deployments directory: %w", err)
	}
	for _, sd := range sds {
		err := generate(sd, defaultDeploymentDir, path.Join(defaultChartDir, sd.chart), tag)
		if err != nil {
			return "", err
		}

		// helm install
		chartPath := path.Join(basePath, defaultDeploymentDir, sd.ServiceName)
		cmdArgs := []string{
			"install", sd.ServiceName, chartPath,
			"-f", path.Join(chartPath, defaultOverridesFile),
			"-n", sd.namespace,
			"--set", "fullnameOverride=" + sd.ServiceName,
		}
		cmd := exec.Command("helm", cmdArgs...)

		out, err := cmd.CombinedOutput()
		output := strings.Replace(string(out), "\n", "\n\t", -1)

		log.Print(string(out))
		if err != nil {
			return "", fmt.Errorf("could not install a helm chart for service %v output: \n%v", sd.ServiceName, output)
		}
	}

	return "", nil
}

func (c Configurd) ServiceDeployments(baseDir, namespace, deployment string) ([]ServiceDeployment, error) {
	file, err := os.Open(path.Join(baseDir, defaultNamespaceDir, namespace, deployment+defaultConfigFileExt))
	defer file.Close()
	if err != nil {
		return []ServiceDeployment{}, err
	}

	var wrapper struct {
		ServiceDeployments []ServiceDeployment `yaml:"services"`
	}
	if err := yaml.NewDecoder(file).Decode(&wrapper); err != nil {
		return []ServiceDeployment{}, fmt.Errorf("could not decode service: %w", err)
	}

	for i, sd := range wrapper.ServiceDeployments {
		svc, err := c.Service(sd.ServiceName)
		if err != nil {
			log.Fatal(err)
		}
		wrapper.ServiceDeployments[i].chart = svc.Chart
		wrapper.ServiceDeployments[i].namespace = namespace
	}
	return wrapper.ServiceDeployments, nil
}
