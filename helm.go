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

type ServiceGroupItem struct {
	ServiceName  string `yaml:"service"`
	Overrides    interface{}
	namespace    string
	serviceGroup string
	chart        string
}

func generateYaml(path string, data interface{}) error {
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

func generate(sgi ServiceGroupItem, outputDir, chartPath string) error {
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

	err = generateYaml(path.Join(dir, defaultOverridesFile), sgi.Overrides)
	if err != nil {
		return fmt.Errorf("could not generate %v: %w", defaultOverridesFile, err)
	}
	return nil
}

func helmInstall(s Service, namespace, outputDir, tag string) error {
	log.Printf("Running helm install for service: %v", s.Name)
	chartPath := path.Join(outputDir, s.Name)
	cmdArgs := []string{
		"install", s.Name, chartPath,
		"-f", path.Join(chartPath, defaultOverridesFile),
		"-n", namespace,
		"--set", "nameOverride=" + s.Name,
		"--set", "fullnameOverride=" + s.Name,
		"--set", "image.repository=" + s.Build.Image,
		"--set", "image.tag=" + tag,
	}
	cmd := exec.Command("helm", cmdArgs...)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)

	if err != nil {
		return fmt.Errorf("could not install a helm chart for service %v output: \n%v", s.Name, output)
	}
	log.Print(string(out))
	return nil
}

func helmUninstall(sgi ServiceGroupItem) error {
	log.Printf("Running helm uninstall for service: %v", sgi.ServiceName)
	cmdArgs := []string{
		"uninstall", "-n", sgi.namespace, sgi.ServiceName,
	}
	cmd := exec.Command("helm", cmdArgs...)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)

	if err != nil {
		return fmt.Errorf("could not uninstall a helm chart for namespace: %v, service group: %v, output: \n%v", sgi.namespace, sgi.serviceGroup, output)
	}
	log.Print(string(out))
	return nil
}

func (c Configurd) Deploy(sgis []ServiceGroupItem, deploymentDir, tag string) error {
	outputDir := path.Join(defaultOutputDir, deploymentDir)

	log.Printf("Removing dir: %v", outputDir)
	if err := os.RemoveAll(outputDir); err != nil {
		return fmt.Errorf("could not clean workdir directory: %w", err)
	}

	log.Printf("Creating dir: %v", outputDir)
	if err := os.MkdirAll(outputDir, 0744); err != nil {
		return fmt.Errorf("could not create a workdir directory: %w", err)
	}
	for _, sgi := range sgis {
		err := generate(sgi, outputDir, path.Join(defaultChartDir, sgi.chart))
		if err != nil {
			return err
		}

	}
	for _, sgi := range sgis {
		s, err := c.Service(sgi.ServiceName)
		if err != nil {
			return err
		}
		err = helmInstall(s, sgi.namespace, outputDir, tag)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c Configurd) Undeploy(sgis []ServiceGroupItem) error {
	for _, sgi := range sgis {
		err := helmUninstall(sgi)
		if err != nil {
			return err
		}
	}
	return nil
}
