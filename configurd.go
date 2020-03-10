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

type ServiceGroup struct {
	name  string
	items []ServiceGroupItem
}

type Namespace struct {
	name          string
	serviceGroups []ServiceGroup
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

func New(log Logger, dir string) (Configurd, error) {
	log.Printf("Parsing configuration files")

	c := Configurd{}
	serviceDir := path.Join(dir, defaultServiceDir)
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
			log.Printf("Loading service: %v, from: %v", name, path)
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
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			log.Printf("Loading namespace: %v, from: %v", info.Name(), path)
			sgs, err := loadServiceGroups(log, sl, info.Name())
			if err != nil {
				return err
			}
			ns := Namespace{
				name:          info.Name(),
				serviceGroups: sgs,
			}
			nss = append(nss, ns)
		}
		return nil
	})
	return nss, err
}

func loadServiceGroups(log Logger, sl serviceLoader, namespace string) ([]ServiceGroup, error) {
	var sgs []ServiceGroup

	sgDir := path.Join(defaultNamespaceDir, namespace)
	err := filepath.Walk(sgDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			var name = info.Name()[0 : len(info.Name())-len(filepath.Ext(info.Name()))]
			log.Printf("Loading service group: %v, from: %v", name, path)
			sgis, err := loadServiceGroupItems(log, sl, namespace, name)
			if err != nil {
				return err
			}
			sg := ServiceGroup{
				name:  name,
				items: sgis,
			}
			sgs = append(sgs, sg)
		}
		return nil
	})
	return sgs, err
}

func loadServiceGroupItems(log Logger, sl serviceLoader, namespace, serviceGroup string) ([]ServiceGroupItem, error) {
	path := path.Join(defaultNamespaceDir, namespace, serviceGroup+defaultConfigFileExt)
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var wrapper struct {
		Items []ServiceGroupItem `yaml:"serviceGroup"`
	}
	if err := yaml.NewDecoder(file).Decode(&wrapper); err != nil {
		return nil, fmt.Errorf("could not decode service group items: %w", err)
	}

	for i, sgi := range wrapper.Items {
		log.Printf("Loading service group item: %v", sgi.ServiceName)
		svc, err := sl(sgi.ServiceName)
		if err != nil {
			return nil, err
		}
		wrapper.Items[i].chart = svc.Chart
		wrapper.Items[i].namespace = namespace
		wrapper.Items[i].serviceGroup = serviceGroup
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

func (c Configurd) GetServiceGroupItems(namespace, serviceGroup string) []ServiceGroupItem {
	var sgis []ServiceGroupItem
	for _, sgi := range c.serviceGroupItems() {
		if sgi.namespace == namespace && (sgi.serviceGroup == serviceGroup || serviceGroup == "") {
			sgis = append(sgis, sgi)
		}
	}
	return sgis
}

func (c Configurd) serviceGroupItems() []ServiceGroupItem {
	var sgis []ServiceGroupItem
	for _, ns := range c.Namespaces {
		for _, sg := range ns.serviceGroups {
			sgis = append(sgis, sg.items...)
		}
	}
	return sgis
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
