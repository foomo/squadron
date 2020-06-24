package configurd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/foomo/config-bob/builder"
	"github.com/foomo/configurd/exampledata"
	"github.com/sirupsen/logrus"

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
	ErrServiceNotFound    = errors.New("service not found")
	ErrBuildNotConfigured = errors.New("build parameter was not configured")
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

type Config struct {
	Tag          string
	BasePath     string
	Log          *logrus.Entry
	TemplateVars TemplateVars
}

type Configurd struct {
	config     Config
	Services   []Service
	Templates  []string
	Namespaces []Namespace
}

type Service struct {
	Name  string `yaml:"-"`
	Image string `yaml:"image"`
	Tag   string `yaml:"tag"`
	Build string `yaml:"build"`
	Chart string `yaml:"chart"`
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
			svc, err := loadService(file, name, config.Tag)
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
	var g Group
	var wrapper struct {
		Group Group `yaml:"group"`
	}
	if err := loadYamlTemplate(path, &wrapper, nil, false); err != nil {
		return wrapper.Group, err
	}
	for name := range wrapper.Group.Services {
		log.Infof("Loading group item: %v", name)
		svc, err := sl(name)
		if err != nil {
			return g, err
		}
		wrapper.Group.Services[name] = loadServiceItem(wrapper.Group.Services[name], svc.Name, namespace, group, svc.Chart)
	}
	wrapper.Group.name = group
	return wrapper.Group, nil
}

func loadServiceItem(item ServiceItem, service, namespace, group, chart string) ServiceItem {
	item.Name = service
	item.namespace = namespace
	item.group = group
	item.chart = chart
	return item
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

func (g Group) ServiceItem(name string) (ServiceItem, error) {
	var available []string
	for _, s := range g.Services {
		if s.Name == name {
			return s, nil
		}
		available = append(available, s.Name)
	}
	return ServiceItem{}, errResourceNotFound(name, "serviceItem", available)
}

func (g Group) ServiceItems() ([]ServiceItem, error) {
	if len(g.Services) == 0 {
		return nil, fmt.Errorf("could not find any service for group: %v", g.name)
	}
	var sis []ServiceItem
	for _, si := range g.Services {
		sis = append(sis, si)
	}
	return sis, nil
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

func loadService(reader io.Reader, name, defaultTag string) (Service, error) {
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

func stringInSlice(str string, slice []string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func isYaml(file string) bool {
	return stringInSlice(filepath.Ext(file), []string{"yml, yaml"})
}

func isJson(file string) bool {
	return filepath.Ext(file) == "json"
}

type TemplateVars map[string]interface{}

func (tv TemplateVars) supportedFileExt() []string {
	return []string{"yml", "yaml", "json"}
}

func NewTemplateVars(workDir string, sourceSlice []string, sourceFile string) (TemplateVars, error) {
	tv := TemplateVars{}
	if err := tv.parseFile(workDir, sourceFile); err != nil {
		return nil, err
	}
	if err := tv.parseSlice(sourceSlice); err != nil {
		return nil, err
	}
	tv["cwd"] = workDir
	return tv, nil
}

func (tv TemplateVars) parseSlice(source []string) error {
	for _, item := range source {
		pieces := strings.Split(item, "=")
		if len(pieces) != 2 || pieces[0] == "" {
			return fmt.Errorf("Invalid format for template var %q, use x=y", item)
		}
		tv[pieces[0]] = pieces[1]
	}
	return nil
}
func (tv TemplateVars) parseFile(workDir, source string) error {
	if source == "" {
		return nil
	}
	if !filepath.IsAbs(source) {
		source = path.Join(workDir, source)
	}
	if !isYaml(source) && !isJson(source) {
		return fmt.Errorf("Unable to parse %q, supported: %v", source, strings.Join(tv.supportedFileExt(), ", "))
	}
	file, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("Error while opening template file: %s", err)
	}
	if isYaml(source) {
		if err := yaml.Unmarshal(file, &tv); err != nil {
			return fmt.Errorf("Error while unmarshalling template file: %s", err)
		}
	}
	if isJson(source) {
		if err := json.Unmarshal(file, &tv); err != nil {
			return fmt.Errorf("Error while unmarshalling template file: %s", err)
		}
		return nil
	}
	return nil
}

func loadYamlTemplate(file string, data interface{}, templateVars interface{}, errOnMissing bool) error {
	tmp, err := template.ParseFiles(file)
	if err != nil {
		return err
	}
	out := bytes.NewBuffer([]byte{})
	if errOnMissing {
		tmp = tmp.Option("missingkey=error")
	}
	if err := tmp.Funcs(builder.TemplateFuncs).Execute(out, templateVars); err != nil {
		return err
	}
	if err := yaml.Unmarshal(out.Bytes(), &data); err != nil {
		return err
	}
	return nil
}
