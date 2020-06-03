package configurd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/foomo/configurd/exampledata"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFileExt   = ".yml"
	defaultServiceDir      = "configurd/services"
	defaultNamespaceDir    = "configurd/namespaces"
	defaultOutputDir       = "configurd/.workdir"
	chartsDir              = "charts"
	chartLockFile          = "Chart.lock"
	chartFile              = "Chart.yaml"
	valuesFile             = "values.yaml"
	defaultChartAPIVersion = "v2"
	defaultChartType       = "application" // application or library
	defaultChartVersion    = "0.1.0"
	defaultChartAppVersion = "1.16.0"
)

var (
	ErrServiceNotFound    = errors.New("service not found")
	ErrBuildNotConfigured = errors.New("build parameter was not configured")
)

type Override map[string]interface{}

type Group struct {
	name     string
	Services map[string]Override
	Jobs     map[string]Override
}

type Namespace struct {
	name   string
	groups []Group
}

type Config struct {
	Tag      string
	BasePath string
	Log      *logrus.Entry
}

type Configurd struct {
	config     Config
	Services   []Service
	Templates  []string
	Namespaces []Namespace
}

type Service struct {
	Name  string          `yaml:"-"`
	Image string          `yaml:"image"`
	Tag   string          `yaml:"tag"`
	Build string          `yaml:"build"`
	Chart ChartDependency `yaml:"chart"`
}

type serviceLoader func(string) (Service, error)

func relativePath(path, basePath string) string {
	return strings.Replace(path, basePath+"/", "", -1)
}

func New(config Config) (Configurd, error) {
	log := config.Log
	log.Infof("Parsing configuration files")
	log.Infof("Entering dir: %q", config.BasePath)

	c := Configurd{
		config: config,
	}

	serviceDir := path.Join(config.BasePath, defaultServiceDir)
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, defaultConfigFileExt) {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			name := strings.TrimSuffix(info.Name(), defaultConfigFileExt)
			log.Infof("Loading service: %v, from: %q", name, relativePath(path, config.BasePath))
			svc, err := loadService(file, name, config.Tag, config.BasePath)
			if err != nil {
				return err
			}
			c.Services = append(c.Services, svc)
		}
		return nil
	})

	if err != nil {
		return Configurd{}, err
	}

	c.Namespaces, err = loadNamespaces(log, c.Service, config.BasePath)

	if err != nil {
		return Configurd{}, err
	}

	return c, nil
}

func loadNamespaces(log *logrus.Entry, sl serviceLoader, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			log.Infof("Loading namespace: %v, from: %q", info.Name(), relativePath(path, basePath))
			gs, err := loadGroups(log, sl, basePath, info.Name())
			if err != nil {
				return err
			}
			ns := Namespace{
				name:   info.Name(),
				groups: gs,
			}
			nss = append(nss, ns)
		}
		return nil
	})
	return nss, err
}

func loadGroups(log *logrus.Entry, sl serviceLoader, basePath, namespace string) ([]Group, error) {
	var gs []Group
	groupPath := path.Join(basePath, defaultNamespaceDir, namespace)
	err := filepath.Walk(groupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			name := strings.TrimSuffix(info.Name(), defaultConfigFileExt)
			log.Infof("Loading group: %v, from: %q", name, relativePath(path, basePath))
			g, err := loadGroup(log, sl, path, namespace, name)
			if err != nil {
				return err
			}
			gs = append(gs, g)
		}
		return nil
	})
	return gs, err
}

func loadGroup(log *logrus.Entry, sl serviceLoader, path, namespace, group string) (Group, error) {
	file, err := os.Open(path)
	if err != nil {
		return Group{}, err
	}
	defer file.Close()

	var wrapper struct {
		Group Group `yaml:"group"`
	}
	if err := yaml.NewDecoder(file).Decode(&wrapper); err != nil {
		return Group{}, fmt.Errorf("could not decode group: %w", err)
	}
	wrapper.Group.name = group
	return wrapper.Group, nil
}

func (c Configurd) Service(name string) (Service, error) {
	var available []string
	for _, s := range c.Services {
		if s.Name == name {
			return s, nil
		}
		available = append(available, s.Name)
	}
	return Service{}, errResourceNotFound(name, "service", available)
}

func (c Configurd) Namespace(name string) (Namespace, error) {
	var available []string
	for _, ns := range c.Namespaces {
		if ns.name == name {
			return ns, nil
		}
		available = append(available, ns.name)
	}
	return Namespace{}, errResourceNotFound(name, "namespace", available)
}

func (ns Namespace) Group(name string) (Group, error) {
	var available []string
	for _, g := range ns.groups {
		if g.name == name {
			return g, nil
		}
		available = append(available, g.name)
	}
	return Group{}, errResourceNotFound(name, "group", available)
}

func (g Group) ServiceOverride(name string) (Override, error) {
	var available []string
	for n, v := range g.Services {
		if name == n {
			return v, nil
		}
		available = append(available, name)
	}
	return nil, errResourceNotFound(name, "serviceOverride", available)
}

func (g Group) ServiceOverrides() (map[string]Override, error) {
	if len(g.Services) == 0 {
		return nil, fmt.Errorf("could not find any service for group: %v", g.name)
	}
	return g.Services, nil
}

func (c Configurd) Build(s Service) (string, error) {
	l := c.config.Log
	if s.Build == "" {
		return "", ErrBuildNotConfigured
	}

	args := strings.Split(s.Build, " ")
	if args[0] == "docker" {
		args = append(strings.Split(s.Build, " "), "-t", fmt.Sprintf("%v:%v", s.Image, s.Tag))
	}
	l.Infof("Building service: %v", s.Name)
	env := []string{
		fmt.Sprintf("TAG=%s", s.Tag),
	}

	return runCommand(l, c.config.BasePath, env, args...)
}

func (c Configurd) Push(name string) (string, error) {
	l := c.config.Log
	s, err := c.Service(name)
	if err != nil {
		return "", fmt.Errorf("could not find service: %w", err)
	}
	image := fmt.Sprintf("%s:%s", s.Image, s.Tag)

	l.Infof("Pushing service %v to %s", s.Name, image)
	return runCommand(l, c.config.BasePath, nil, "docker", "push", image)
}

func loadService(reader io.Reader, name, defaultTag, basePath string) (Service, error) {
	var wrapper struct {
		Service Service `yaml:"service"`
	}
	if err := yaml.NewDecoder(reader).Decode(&wrapper); err != nil {
		return Service{}, fmt.Errorf("could not decode service: %w", err)
	}
	wrapper.Service.Name = name
	if wrapper.Service.Tag == "" {
		wrapper.Service.Tag = defaultTag
	}
	wrapper.Service.Chart.Repository =
		strings.Replace(wrapper.Service.Chart.Repository, "file://./", fmt.Sprintf("file://%v/", basePath), 1)
	// correct the relative path for the file:// chart repository
	wrapper.Service.Chart.Alias = name
	return wrapper.Service, nil
}

func Init(log *logrus.Logger, dir string) (string, error) {
	log.Infof("Downloading example configuration into dir: %q", dir)
	return "", exampledata.RestoreAssets(dir, "")
}

func runCommand(log *logrus.Entry, cwd string, env []string, command ...string) (string, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = cwd
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, env...)

	log.Tracef("executing %q from wd %q", cmd.String(), cmd.Dir)

	cmdReader, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		return "", err
	}

	var out []string
	scanner := bufio.NewScanner(cmdReader)
	for scanner.Scan() {
		line := scanner.Text()
		log.Trace(line)
		out = append(out, line)
	}
	output := strings.Join(out, "\n")
	if err := cmd.Wait(); err != nil {
		return output, fmt.Errorf("%s, %s", err.Error(), output)
	}

	return output, err
}

func errResourceNotFound(name, resource string, available []string) error {
	if name == "" {
		return fmt.Errorf("%s not provided. Available: %s", resource, strings.Join(available, ", "))
	}
	return fmt.Errorf("%s '%s' not found. Available: %s", resource, name, strings.Join(available, ", "))
}
