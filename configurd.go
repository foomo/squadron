package configurd

import (
	"errors"
	"fmt"
	"io"
	"log"
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
	defaultDeploymentDir = "configurd/deployments"
	defaultOverridesFile = "overrides.yaml"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

type Deployment struct {
	name               string
	serviceDeployments []ServiceDeployment
}

type Namespace struct {
	name        string
	deployments []Deployment
}

type Configurd struct {
	Services   []Service
	Templates  []string
	Namespaces []Namespace
}

func New(dir string) (Configurd, error) {
	c := Configurd{}

	//load services
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
			svc, err := loadService(file)
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

	c.Namespaces, err = loadNamespaces(c, dir)

	if err != nil {
		return Configurd{}, err
	}

	return c, nil
}

func loadNamespaces(c Configurd, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			ds, err := loadDeployments(c, info.Name())
			if err != nil {
				return err
			}
			ns := Namespace{
				name:        info.Name(),
				deployments: ds,
			}
			nss = append(nss, ns)
		}
		return nil
	})
	return nss, err
}

func loadDeployments(c Configurd, namespace string) ([]Deployment, error) {
	var ds []Deployment
	err := filepath.Walk(path.Join(defaultNamespaceDir, namespace), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			var name = info.Name()[0 : len(info.Name())-len(filepath.Ext(info.Name()))]
			sds, err := loadServiceDeployments(c, namespace, name)
			if err != nil {
				return err
			}
			d := Deployment{
				name:               name,
				serviceDeployments: sds,
			}
			ds = append(ds, d)
		}
		return nil
	})
	return ds, err
}

func loadServiceDeployments(c Configurd, namespace, deployment string) ([]ServiceDeployment, error) {
	path := path.Join(defaultNamespaceDir, namespace, deployment+defaultConfigFileExt)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		return []ServiceDeployment{}, err
	}

	var wrapper struct {
		ServiceDeployments []ServiceDeployment `yaml:"services"`
	}
	if err := yaml.NewDecoder(file).Decode(&wrapper); err != nil {
		return []ServiceDeployment{}, fmt.Errorf("could not decode service deployment: %w", err)
	}

	for i, sd := range wrapper.ServiceDeployments {
		svc, err := c.Service(sd.ServiceName)
		if err != nil {
			log.Fatal(err)
		}
		wrapper.ServiceDeployments[i].chart = svc.Chart
		wrapper.ServiceDeployments[i].namespace = namespace
		wrapper.ServiceDeployments[i].deployment = deployment
	}
	return wrapper.ServiceDeployments, nil
}

func (c Configurd) Service(name string) (Service, error) {
	for _, svc := range c.Services {
		if svc.Name == name {
			return svc, nil
		}
	}
	return Service{}, ErrServiceNotFound
}

func (c Configurd) GetServiceDeployments(namespace, deployment string) []ServiceDeployment {
	var sds []ServiceDeployment
	for _, sd := range c.serviceDeployments() {
		if sd.namespace == namespace && sd.deployment == deployment {
			sds = append(sds, sd)
		}
	}
	return sds
}

func (c Configurd) serviceDeployments() []ServiceDeployment {
	var sds []ServiceDeployment
	for _, ns := range c.Namespaces {
		for _, d := range ns.deployments {
			sds = append(sds, d.serviceDeployments...)
		}
	}
	return sds
}

func loadService(reader io.Reader) (Service, error) {
	var wrapper struct {
		Service Service `yaml:"service"`
	}
	if err := yaml.NewDecoder(reader).Decode(&wrapper); err != nil {
		return Service{}, fmt.Errorf("could not decode service: %w", err)
	}
	return wrapper.Service, nil
}
