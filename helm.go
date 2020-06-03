package configurd

import (
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/mitchellh/mapstructure"
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

type Image struct {
	Repository string
	Tag        string
}

type Volume struct {
	Name  string
	Host  string
	Mount string
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

func updateVolumeOverride(basePath string, override Override) (Override, error) {
	if _, ok := override["volumes"]; ok {
		// if there are volume overrides, make relative paths absolute
		var vs []Volume
		if err := mapstructure.Decode(override["volumes"], &vs); err != nil {
			return nil, err
		}
		for i := range vs {
			vs[i].Host = strings.Replace(vs[i].Host, "./", fmt.Sprintf("%v/", basePath), 1)
		}
		override["volumes"] = vs
	}
	return override, nil
}

func updateImageOverride(image, tag string, override Override) (Override, error) {
	if _, ok := override["image"]; !ok {
		// if repository and tag are not in overrides, add them
		override["image"] = Image{image, tag}
	}
	return override, nil
}

func helmUpdateDependency(log *logrus.Entry, group, groupChartPath string) (string, error) {
	log.Infof("Running helm dependency update for group: %v", group)
	cmd := []string{"helm", "dependency", "update", groupChartPath}
	return runCommand(log, "", nil, cmd...)
}

func helmInstall(log *logrus.Entry, group, namespace, groupChartPath string) (string, error) {
	log.Infof("Running helm install for group: %v", group)
	cmd := []string{
		"helm", "upgrade", group, groupChartPath,
		"-n", namespace,
		"--install",
	}
	return runCommand(log, "", nil, cmd...)
}

func helmUninstall(log *logrus.Entry, group, namespace string) (string, error) {
	log.Infof("Running helm uninstall for group: %v", group)
	cmd := []string{
		"helm",
		"uninstall",
		"-n", namespace,
		group,
	}
	return runCommand(log, "", nil, cmd...)
}

type InstallConfiguration struct {
	ServiceOverrides map[string]Override
	BasePath         string
	OutputDir        string
	Tag              string
	Verbose          bool
	Group            string
	Namespace        string
}

func (c Configurd) Install(cnf InstallConfiguration) (string, error) {
	logger := c.config.Log

	logger.Infof("Installing services")
	groupChartPath := path.Join(cnf.BasePath, defaultOutputDir, cnf.OutputDir, cnf.Group)

	logger.Infof("Entering dir: %q", path.Join(cnf.BasePath, defaultOutputDir))
	logger.Printf("Creating dir: %q", cnf.OutputDir)
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

	groupChart := NewChart(cnf.Group)
	for name, override := range cnf.ServiceOverrides {
		s, err := c.Service(name)
		if err != nil {
			return "", err
		}
		override, err = updateVolumeOverride(cnf.BasePath, override)
		if err != nil {
			return "", err
		}
		override, err = updateImageOverride(s.Image, s.Tag, override)
		if err != nil {
			return "", err
		}
		cnf.ServiceOverrides[name] = override
		groupChart.Dependencies = append(groupChart.Dependencies, s.Chart)
	}

	if err := generateYaml(logger, path.Join(groupChartPath, chartFile), groupChart); err != nil {
		return "", err
	}
	if err := generateYaml(logger, path.Join(groupChartPath, valuesFile), cnf.ServiceOverrides); err != nil {
		return "", err
	}

	output, err := helmUpdateDependency(logger, cnf.Group, groupChartPath)
	if err != nil {
		return output, err
	}
	return helmInstall(logger, cnf.Group, cnf.Namespace, groupChartPath)
}

func (c Configurd) Uninstall(group, namespace string, verbose bool) (string, error) {
	logger := c.config.Log

	output, err := helmUninstall(logger, group, namespace)
	if err != nil {
		return output, err
	}
	return output, nil
}
