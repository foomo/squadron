package configurd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFileExt = ".yml"
	defaultServiceDir    = "configurd/services"
	defaultNamespaceDir  = "configurd/namespaces"
	defaultChartDir      = "configurd/charts"
	defaultOutputDir     = "configurd/.workdir"
	defaultOverridesFile = "overrides.yaml"
	defaultInitUrl       = "https://github.com/foomo/configurd.git/branches/feature/helm-charts-deployments/example"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type Group struct {
	name     string
	Services map[string]ServiceItem
	Jobs     map[string]JobItem
}

type Namespace struct {
	name   string
	groups []Group
}

type Configurd struct {
	Services   []Service
	Templates  []string
	Namespaces []Namespace
}

type serviceLoader func(string) (Service, error)

type Logger interface {
	Printf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

func relativePath(path, basePath string) string {
	return strings.Replace(path, basePath+"/", "", -1)
}

func New(log Logger, basePath string) (Configurd, error) {
	log.Printf("Parsing configuration files")
	log.Printf("Entering dir: %q", basePath)

	c := Configurd{}
	serviceDir := path.Join(basePath, defaultServiceDir)
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			var name = info.Name()[0 : len(info.Name())-len(filepath.Ext(info.Name()))]
			log.Printf("Loading service: %v, from: %q", name, relativePath(path, basePath))
			svc, err := loadService(file, name)
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

	c.Namespaces, err = loadNamespaces(log, c.Service, basePath)

	if err != nil {
		return Configurd{}, err
	}

	return c, nil
}

func loadNamespaces(log Logger, sl serviceLoader, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			log.Printf("Loading namespace: %v, from: %q", info.Name(), relativePath(path, basePath))
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

func loadGroups(log Logger, sl serviceLoader, basePath, namespace string) ([]Group, error) {
	var gs []Group
	groupPath := path.Join(basePath, defaultNamespaceDir, namespace)
	err := filepath.Walk(groupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			var name = info.Name()[0 : len(info.Name())-len(filepath.Ext(info.Name()))]
			log.Printf("Loading group: %v, from: %q", name, relativePath(path, basePath))
			g, err := loadGroup(log, sl, path, namespace, name)
			if err != nil {
				return err
			}
			g.name = name
			gs = append(gs, g)
		}
		return nil
	})
	return gs, err
}

func loadGroup(log Logger, sl serviceLoader, path, namespace, group string) (Group, error) {
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

	for name := range wrapper.Group.Services {
		log.Printf("Loading service group item: %v", name)
		svc, err := sl(name)
		if err != nil {
			return Group{}, err
		}
		wrapper.Group.Services[name] = loadServiceItem(wrapper.Group.Services[name], svc.Name, svc.Chart, namespace, group)
	}
	return wrapper.Group, nil
}

func loadServiceItem(item ServiceItem, service, chart, namespace, group string) ServiceItem {
	item.ServiceName = service
	item.chart = chart
	item.namespace = namespace
	item.group = group
	return item
}

func (c Configurd) Service(name string) (Service, error) {
	for _, svc := range c.Services {
		if svc.Name == name {
			return svc, nil
		}
	}
	return Service{}, ErrServiceNotFound
}

func (c Configurd) GetServiceItems(namespace, group string) []ServiceItem {
	var sis []ServiceItem
	for _, si := range c.serviceItems() {
		if si.namespace == namespace && si.group == group {
			sis = append(sis, si)
		}
	}
	return sis
}

func (c Configurd) serviceItems() []ServiceItem {
	var sis []ServiceItem
	for _, ns := range c.Namespaces {
		for _, g := range ns.groups {
			for _, si := range g.Services {
				sis = append(sis, si)
			}
		}
	}
	return sis
}

func loadService(reader io.Reader, name string) (Service, error) {
	var wrapper struct {
		Service Service `yaml:"service"`
	}
	if err := yaml.NewDecoder(reader).Decode(&wrapper); err != nil {
		return Service{}, fmt.Errorf("could not decode service: %w", err)
	}
	wrapper.Service.Name = name
	return wrapper.Service, nil
}

func logOutput(log Logger, verbose bool, format string, args ...interface{}) {
	if verbose {
		log.Printf(format, args...)
	}
}

func Init(log Logger, dir string, flagVerbose bool) (string, error) {
	// log.Printf("Creating dir: %q", dir)

	// if err := os.MkdirAll(dir, 0744); err != nil {
	// 	return "", fmt.Errorf("could not create a directory: %w", err)
	// }

	log.Printf("Downloading example configuration into dir: %q", dir)
	cmd := exec.Command("svn", "export", defaultInitUrl, dir)

	out, err := cmd.CombinedOutput()
	output := strings.Replace(string(out), "\n", "\n\t", -1)
	if err != nil {
		return "", fmt.Errorf("could not download a configurd example, output: \n%v", output)
	}
	return output, nil
}
