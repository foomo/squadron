package squadron

import (
	"fmt"
	"os"
	"path"

	"github.com/sirupsen/logrus"
)

type ChartDependency struct {
	Name       string
	Repository string `yaml:"repository,omitempty"`
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

func helmUpdateDependency(l *logrus.Entry, group, groupChartPath string) (string, error) {
	l.Infof("Running helm dependency update for group: %v", group)
	cmd := []string{"helm", "dependency", "update", groupChartPath}
	return Command(l, cmd...).Run()
}

func helmInstall(l *logrus.Entry, group, namespace, groupChartPath string) (string, error) {
	l.Infof("Running helm install for group: %v", group)
	cmd := []string{
		"helm", "upgrade", group, groupChartPath,
		"-n", namespace,
		"--install",
	}
	return Command(l, cmd...).Run()
}

func helmUninstall(l *logrus.Entry, group, namespace string) (string, error) {
	l.Infof("Running helm uninstall for group: %v", group)
	cmd := []string{
		"helm",
		"uninstall",
		"-n", namespace,
		group,
	}
	return Command(l, cmd...).Run()
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

func (c Squadron) Install(ors map[string]Override, basePath, outputDir, namespace, group, tag string) (string, error) {
	c.l.Infof("Installing services")
	groupChartPath := path.Join(basePath, defaultOutputDir, outputDir, group)

	c.l.Infof("Entering dir: %q", path.Join(basePath, defaultOutputDir))
	c.l.Printf("Creating dir: %q", outputDir)
	if _, err := os.Stat(groupChartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(groupChartPath, 0744); err != nil {
			return "", fmt.Errorf("could not create a workdir directory: %w", err)
		}
	}

	chartsPath := path.Join(groupChartPath, chartsDir)
	c.l.Infof("Removing dir: %q", chartsPath)
	if err := os.RemoveAll(chartsPath); err != nil {
		return "", fmt.Errorf("could not clean charts directory: %w", err)
	}
	groupChartLockPath := path.Join(groupChartPath, chartLockFile)
	c.l.Infof("Removing file: %q", groupChartLockPath)
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

	if err := GenerateYaml(path.Join(groupChartPath, chartFile), groupChart); err != nil {
		return "", err
	}
	if err := GenerateYaml(path.Join(groupChartPath, valuesFile), ors); err != nil {
		return "", err
	}

	output, err := helmUpdateDependency(c.l, group, groupChartPath)
	if err != nil {
		return output, err
	}

	return helmInstall(c.l, group, namespace, groupChartPath)
}

func (c Squadron) Uninstall(group, namespace string) (string, error) {
	output, err := helmUninstall(c.l, group, namespace)
	if err != nil {
		return output, err
	}
	return output, nil
}
