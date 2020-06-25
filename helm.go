package configurd

import (
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type ChartDependency struct {
	Name       string
	Repository string
	Version    string
	Alias      string
}

type Chart struct {
	APIVersion   string `yaml:"apiVersion"`
	Name         string
	Description  string
	Type         string
	Version      string
	AppVersion   string `yaml:"appVersion"`
	Dependencies []ChartDependency
}

func NewChart(name string) *Chart {
	return &Chart{
		APIVersion:  defaultChartAPIVersion,
		Name:        name,
		Description: fmt.Sprintf("A helm parent chart for group %v", name),
		Type:        defaultChartType,
		Version:     defaultChartVersion,
		AppVersion:  defaultChartAppVersion,
	}
}

type JobItem struct {
	Name      string
	Overrides interface{}
	namespace string
	group     string
	chart     string
}

func generateYaml(_ *logrus.Entry, path string, data interface{}) error {
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

func helmUpdateDependency(l *logrus.Entry, group, groupChartPath string) (string, error) {
	l.Infof("Running helm dependency update for group: %v", group)
	cmd := []string{"helm", "dependency", "update", groupChartPath}
	return command(l, cmd...).run()
}

func helmInstall(l *logrus.Entry, group, namespace, groupChartPath string) (string, error) {
	l.Infof("Running helm install for group: %v", group)
	cmd := []string{
		"helm", "upgrade", group, groupChartPath,
		"-n", namespace,
		"--install",
	}
	return command(l, cmd...).run()
}

func helmUninstall(l *logrus.Entry, group, namespace string) (string, error) {
	l.Infof("Running helm uninstall for group: %v", group)
	cmd := []string{
		"helm",
		"uninstall",
		"-n", namespace,
		group,
	}
	return command(l, cmd...).run()
}

func updateImageOverride(image, tag string, override Override) (Override, error) {
	if _, ok := override["image"]; !ok {
		// if repository and tag are not in overrides, add them
		override["image"] = map[string]string{
			"repository": image,
			"tag":        tag,
		}
	}
	return override, nil
}

func (c Configurd) Install(ors map[string]Override, basePath, outputDir, namespace, group, tag string) (string, error) {
	logger := c.config.Log

	logger.Infof("Installing services")
	groupChartPath := path.Join(basePath, defaultOutputDir, outputDir, group)

	logger.Infof("Entering dir: %q", path.Join(basePath, defaultOutputDir))
	logger.Printf("Creating dir: %q", outputDir)
	if _, err := os.Stat(groupChartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(groupChartPath, 0744); err != nil {
			return "", fmt.Errorf("could not create a workdir directory: %w", err)
		}
	}

	chartsPath := path.Join(groupChartPath, chartsDir)
	logger.Infof("Removing dir: %q", chartsPath)
	if err := os.RemoveAll(chartsPath); err != nil {
		return "", fmt.Errorf("could not clean charts directory: %w", err)
	}
	groupChartLockPath := path.Join(groupChartPath, chartLockFile)
	logger.Infof("Removing file: %q", groupChartLockPath)
	if err := os.RemoveAll(groupChartLockPath); err != nil {
		return "", fmt.Errorf("could not clean workdir directory: %w", err)
	}

	groupChart := NewChart(group)
	for name, override := range ors {
		s, err := c.Service(name)
		if err != nil {
			return "", err
		}
		ors[name], err = updateImageOverride(s.Image, s.Tag, override)
		if err != nil {
			return "", err
		}
		groupChart.Dependencies = append(groupChart.Dependencies, s.Chart)
	}

	if err := generateYaml(logger, path.Join(groupChartPath, chartFile), groupChart); err != nil {
		return "", err
	}
	if err := generateYaml(logger, path.Join(groupChartPath, valuesFile), ors); err != nil {
		return "", err
	}

	output, err := helmUpdateDependency(logger, group, groupChartPath)
	if err != nil {
		return output, err
	}

	return helmInstall(logger, group, namespace, groupChartPath)
}

func (c Configurd) Uninstall(group, namespace string) (string, error) {
	logger := c.config.Log

	output, err := helmUninstall(logger, group, namespace)
	if err != nil {
		return output, err
	}
	return output, nil
}
