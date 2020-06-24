package configurd

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mitchellh/mapstructure"
	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Volume struct {
	Name  string
	Host  string
	Mount string
}

type ServiceItem struct {
	Name      string
	overrides interface{}
	namespace string
	group     string
	chart     string
}

func (si ServiceItem) getOverrides(basePath string, tv TemplateVars) (interface{}, error) {
	if si.overrides == nil {
		path := path.Join(basePath, defaultNamespaceDir, si.namespace, si.group, defaultConfigFileExt)
		var wrapper struct {
			Group Group `yaml:"group"`
		}
		if err := loadYamlTemplate(path, &wrapper, tv, true); err != nil {
			return nil, err
		}
		si.overrides = wrapper.Group.Services[si.Name].overrides
	}
	return si.overrides, nil
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

func fixVolumeRelativePath(basePath string, volumes interface{}) []Volume {
	var vs []Volume
	mapstructure.Decode(volumes, &vs)
	for i := 0; i < len(vs); i++ {
		vs[i].Host = strings.Replace(vs[i].Host, "./", fmt.Sprintf("%v/", basePath), 1)
	}
	return vs
}

func generate(log *logrus.Entry, si ServiceItem, basePath, outputDir string, tv TemplateVars) error {
	outputPath := path.Join(basePath, defaultOutputDir, outputDir, si.Name)
	log.Infof("Creating dir: %q", path.Join(outputDir, si.Name))
	if err := os.MkdirAll(outputPath, 0744); err != nil {
		return fmt.Errorf("could not create output dir: %w", err)
	}

	log.Infof("Copying chart: %v to dir: %q", si.chart, path.Join(outputDir, si.Name))
	chartPath := path.Join(basePath, defaultChartDir, si.chart)
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

	overrides, err := si.getOverrides(basePath, tv)
	if err != nil {
		return err
	}

	log.Printf("Generating yaml file: %q", path.Join(outputDir, si.Name, defaultOverridesFile))
	err = generateYaml(log, path.Join(outputPath, defaultOverridesFile), overrides)
	if err != nil {
		return fmt.Errorf("could not generate %v: %w", defaultOverridesFile, err)
	}
	return nil
}

func helmInstall(log *logrus.Entry, si ServiceItem, service Service, outputDir string) (string, error) {
	log.Infof("Running helm install for service: %v", si.Name)

	chartPath := path.Join(outputDir, si.Name)
	cmd := []string{
		"helm", "upgrade", si.Name, chartPath,
		"-f", path.Join(chartPath, defaultOverridesFile),
		"-n", si.namespace,
		"--install",
		"--set", fmt.Sprintf("group=%v", si.group),
		"--set", fmt.Sprintf("metadata.name=%s", service.Name),
		"--set", fmt.Sprintf("metadata.component=%s", si.group),
		"--set", fmt.Sprintf("metadata.namespace=%s", si.namespace),
	}
	if service.Image != "" {
		cmd = append(cmd, "--set", fmt.Sprintf("image.repository=%s", service.Image))
	}
	if service.Tag != "" {
		cmd = append(cmd, "--set", fmt.Sprintf("image.tag=%s", service.Tag))
	}
	return runCommand(log, "", nil, cmd...)
}

func helmUninstall(log *logrus.Entry, si ServiceItem) (string, error) {
	log.Infof("Running helm uninstall for service: %v", si.Name)
	cmd := []string{
		"helm",
		"uninstall",
		"-n", si.namespace,
		si.Name,
	}
	return runCommand(log, "", nil, cmd...)
}

type InstallConfiguration struct {
	ServiceItems []ServiceItem
	BasePath     string
	OutputDir    string
	Tag          string
	TemplateVars map[string]interface{}
	Verbose      bool
}

func (c Configurd) Install(cnf InstallConfiguration) (string, error) {
	logger := c.config.Log

	logger.Infof("Installing services")
	outputPath := path.Join(cnf.BasePath, defaultOutputDir, cnf.OutputDir)

	logger.Infof("Entering dir: %q", path.Join(cnf.BasePath, defaultOutputDir))
	logger.Infof("Removing dir: %q", cnf.OutputDir)
	if err := os.RemoveAll(outputPath); err != nil {
		return "", fmt.Errorf("could not clean workdir directory: %w", err)
	}

	logger.Printf("Creating dir: %q", cnf.OutputDir)
	if err := os.MkdirAll(outputPath, 0744); err != nil {
		return "", fmt.Errorf("could not create a workdir directory: %w", err)
	}
	for _, si := range cnf.ServiceItems {
		err := generate(logger, si, cnf.BasePath, cnf.OutputDir, cnf.TemplateVars)
		if err != nil {
			return "", err
		}
	}

	var output []string
	for _, si := range cnf.ServiceItems {
		s, err := c.Service(si.Name)
		if err != nil {
			return "", err
		}
		out, err := helmInstall(logger, si, s, outputPath)
		if err != nil {
			return out, err
		}
		output = append(output, out)
	}

	return strings.Join(output, "\n"), nil
}

func (c Configurd) Uninstall(sis []ServiceItem, namespace string) (string, error) {
	logger := c.config.Log

	var outputs []string
	for _, si := range sis {
		out, err := helmUninstall(logger, si)
		if err != nil {
			return out, err
		}
		outputs = append(outputs, out)
	}
	return strings.Join(outputs, "\n"), nil
}
