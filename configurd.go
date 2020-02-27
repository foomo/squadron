package configurd

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	defaultServiceDir = "configurd/services"
)

var (
	ErrServiceNotFound = errors.New("service not found")
)

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
		if !info.IsDir() && strings.HasSuffix(path, ".yml") {
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

	if err != nil {
		return Configurd{}, err
	}

	// Load Templates
	// Load Namespaces

	// Validate Connections
	// - Services to templates
	// - Namespaces to services
	return c, nil
}

func (c Configurd) Service(name string) (Service, error) {
	for _, svc := range c.Services {
		if svc.Name == name {
			return svc, nil
		}
	}
	return Service{}, ErrServiceNotFound
}
