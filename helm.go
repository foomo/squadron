package configurd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
	"gopkg.in/yaml.v2"
)

type ServiceDeployment struct {
	ServiceName string `yaml:"service"`
	Overrides   interface{}
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

	err = generateYaml(path.Join(dir, "overrides.yaml"), sd.Overrides)
	if err != nil {
		return fmt.Errorf("could not generate overrides.yaml: %w", err)
	}
	return nil
}

func (c Configurd) Deploy(sd ServiceDeployment, namespace, tag string) (string, error) {
	output := "deployments"
	if err := os.RemoveAll(output); err != nil {
		return "", fmt.Errorf("could not clean deployment directory: %w", err)
	}

	if err := os.MkdirAll(output, 0744); err != nil {
		return "", fmt.Errorf("could not create a deployments directory: %w", err)
	}

	err := generate(sd, output, "configurd/charts/example", tag)
	if err != nil {
		return "", err
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
	return wrapper.ServiceDeployments, nil
}
