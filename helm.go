package configurd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
)

type ServiceDeployment struct {
	Name        string
	ServiceName string `yaml:"service"`
	Overrides   interface{}
}

func generate(deployment, outputDir, chartPath string) error {
	dir := path.Join(outputDir, deployment)
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

	//TODO: Write overrides into overrides.yaml (global)

	return nil
}

func (c Configurd) Deploy(namespace, deployment, tag string) (string, error) {
	//TODO: Generate Output YAML file
	output := "deployments"
	if err := os.RemoveAll(output); err != nil {
		return "", fmt.Errorf("coudl not clean deployment directory: %w", err)
	}

	if err := os.MkdirAll("deployment", 0744); err != nil {
		return "", fmt.Errorf("could not create a deployment directory: %w", err)
	}

	err := generate(deployment, output, "configurd/charts/example")
	if err != nil {
		return "", err
	}

	return "", nil
}
