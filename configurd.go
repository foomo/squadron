package configurd

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	defaultConfigFileExt      = ".yml"
	defaultServiceDir         = "configurd/services"
	defaultTemplateServiceDir = "configurd/templates/services"
	defaultNamespaceDir       = "configurd/namespaces"
)

func ErrResourceNotFound(resourceType string, resourceName string) error {
	return errors.New(resourceType + " resource named " + resourceName + " not found")
}

type Namespace struct {
	Name string
}

type Configurd struct {
	Services   []Service
	Templates  []string
	Namespaces []Namespace
}

func New(dir string) (Configurd, error) {
	c := Configurd{}

	serviceDir := path.Join(dir, defaultServiceDir)
	err := filepath.Walk(serviceDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			svc, err := LoadService(file)
			if err != nil {
				return err
			}
			c.Services = append(c.Services, svc)
		}
		return nil
	})

	// Load Templates
	err = filepath.Walk(path.Join(defaultTemplateServiceDir), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			c.Templates = append(c.Templates, strings.TrimSuffix(info.Name(), defaultConfigFileExt))
		}
		return nil
	})

	// Load Namespaces
	namespaceDir := path.Join(dir, defaultNamespaceDir)
	err = filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() && path != namespaceDir {
			c.Namespaces = append(c.Namespaces, Namespace{Name: info.Name()})
		}
		return nil
	})

	if err != nil {
		return Configurd{}, err
	}

	// Validate Connections
	// - Services to templates
	for _, service := range c.Services {
		if !c.ServiceHasTemplate(service.Name) {
			return Configurd{}, ErrResourceNotFound("service template", service.Name)
		}
	}

	// - Namespaces to services
	for _, namespace := range c.Namespaces {
		for _, service := range c.Services {
			if !c.NamespaceHasService(dir, namespace.Name, service.Name) {
				return Configurd{}, ErrResourceNotFound("namespace service", service.Name)
			}
		}
	}

	return c, nil
}

func (c Configurd) Service(name string) (Service, error) {
	for _, svc := range c.Services {
		if svc.Name == name {
			return svc, nil
		}
	}
	return Service{}, ErrResourceNotFound("service", name)
}

func (c Configurd) ServiceHasTemplate(serviceName string) bool {
	for _, t := range c.Templates {
		if t == serviceName {
			return true
		}
	}
	return false
}

func (c Configurd) NamespaceHasService(basePath string, namespaceName string, serviceName string) bool {
	path := path.Join(basePath, defaultNamespaceDir, namespaceName, serviceName+defaultConfigFileExt)
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}
