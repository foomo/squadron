package configurd

import (
	"errors"
	"fmt"
	"io"
	"os"
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
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type Group struct {
	name     string
	Services []ServiceItem
	Jobs     []JobItem
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

func relativePath(absoltePath, basePath string) string {
	return strings.Replace(absoltePath, basePath+"/", "", -1)
}

func New(log Logger, dir string) (Configurd, error) {
	log.Printf("Parsing configuration files")

	c := Configurd{}
	serviceDir := path.Join(dir, defaultServiceDir)
	log.Printf("Entering dir: %q", serviceDir)
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
			log.Printf("Loading service: %v, from file: %q", name, relativePath(path, serviceDir))
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

	c.Namespaces, err = loadNamespaces(log, c.Service, dir)

	if err != nil {
		return Configurd{}, err
	}

	return c, nil
}

func loadNamespaces(log Logger, sl serviceLoader, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	log.Printf("Entering dir: %q", namespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			log.Printf("Loading namespace: %v, from dir: %q", info.Name(), relativePath(path, namespaceDir))
			sgs, err := loadGroups(log, sl, basePath, info.Name())
			if err != nil {
				return err
			}
			ns := Namespace{
				name:   info.Name(),
				groups: sgs,
			}
			nss = append(nss, ns)
		}
		return nil
	})
	return nss, err
}

func loadGroups(log Logger, sl serviceLoader, basePath, namespace string) ([]Group, error) {
	var sgs []Group

	groupDir := path.Join(basePath, defaultNamespaceDir, namespace)
	log.Printf("Entering dir: %q", groupDir)
	err := filepath.Walk(groupDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			var name = info.Name()[0 : len(info.Name())-len(filepath.Ext(info.Name()))]
			log.Printf("Loading service group: %v, from file: %q", name, relativePath(path, groupDir))
			sis, err := loadServiceItems(log, sl, namespace, name)
			if err != nil {
				return err
			}
			sg := Group{
				name:     name,
				Services: sis,
			}
			sgs = append(sgs, sg)
		}
		return nil
	})
	return sgs, err
}

func loadServiceItems(log Logger, sl serviceLoader, namespace, group string) ([]ServiceItem, error) {
	path := path.Join(defaultNamespaceDir, namespace, group+defaultConfigFileExt)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var wrapper struct {
		Items []ServiceItem `yaml:"group"`
	}
	if err := yaml.NewDecoder(file).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("could not decode service group items: %w", err)
	}

	for i, si := range wrapper.Items {
		log.Printf("Loading service group item: %v", si.ServiceName)
		svc, err := sl(si.ServiceName)
		if err != nil {
			return nil, err
		}
		wrapper.Items[i].chart = svc.Chart
		wrapper.Items[i].namespace = namespace
		wrapper.Items[i].group = group
	}
	return wrapper.Items, nil
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
		if si.namespace == namespace && (si.group == group || group == "") {
			sis = append(sis, si)
		}
	}
	return sis
}

func (c Configurd) serviceItems() []ServiceItem {
	var sis []ServiceItem
	for _, ns := range c.Namespaces {
		for _, sg := range ns.groups {
			sis = append(sis, sg.Services...)
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
