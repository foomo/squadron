package squadron

import (
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
	"time"

	"github.com/foomo/config-bob/builder"
	"github.com/foomo/squadron/exampledata"
	"github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

const (
	defaultConfigFileExt   = ".yml"
	defaultServiceDir      = "squadron/services"
	defaultNamespaceDir    = "squadron/namespaces"
	defaultOutputDir       = "squadron/.workdir"
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

type Squadron struct {
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

func New(config Config) (Squadron, error) {
	l := config.Log
	l.Infof("Parsing configuration files")
	l.Infof("Entering dir: %q", config.BasePath)

	c := Squadron{
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
			l.Infof("Loading service: %v, from: %q", name, relativePath(path, config.BasePath))
			svc, err := loadService(file, name, config.Tag, config.BasePath)
			if err != nil {
				return err
			}
			c.Services = append(c.Services, svc)
		}
		return nil
	})

	if err != nil {
		return Squadron{}, err
	}

	c.Namespaces, err = loadNamespaces(l, c.Service, config.BasePath)

	if err != nil {
		return Squadron{}, err
	}

	return c, nil
}

func loadNamespaces(l *logrus.Entry, sl serviceLoader, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			l.Infof("Loading namespace: %v, from: %q", info.Name(), relativePath(path, basePath))
			gs, err := loadGroups(l, sl, basePath, info.Name())
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

func loadGroups(l *logrus.Entry, sl serviceLoader, basePath, namespace string) ([]Group, error) {
	var gs []Group
	groupPath := path.Join(basePath, defaultNamespaceDir, namespace)
	err := filepath.Walk(groupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			name := strings.TrimSuffix(info.Name(), defaultConfigFileExt)
			l.Infof("Loading group: %v, from: %q", name, relativePath(path, basePath))
			g, err := loadGroup(l, sl, path, namespace, name)
			if err != nil {
				return err
			}
			gs = append(gs, g)
		}
		return nil
	})
	return gs, err
}

func loadGroup(l *logrus.Entry, sl serviceLoader, path, namespace, group string) (Group, error) {
	var wrapper struct {
		Group Group `yaml:"group"`
	}
	bs, err := parseTemplate(path, nil, false)
	if err != nil {
		return wrapper.Group, err
	}
	if err := yaml.Unmarshal(bs, &wrapper); err != nil {
		return wrapper.Group, err
	}
	wrapper.Group.name = group
	for name := range wrapper.Group.Services {
		// the overrides have not been parsed with templates
		// we only need this on install
		// so use nil instead of wrong values
		wrapper.Group.Services[name] = nil
	}
	return wrapper.Group, nil
}

func (c Squadron) Service(name string) (Service, error) {
	var available []string
	for _, s := range c.Services {
		if s.Name == name {
			return s, nil
		}
		available = append(available, s.Name)
	}
	return Service{}, errResourceNotFound(name, "service", available)
}

func (c Squadron) Namespace(name string) (Namespace, error) {
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

func (g Group) Overrides(basePath, namespace string, tv TemplateVars) (map[string]Override, error) {
	path := path.Join(basePath, defaultNamespaceDir, namespace, g.name+defaultConfigFileExt)
	var wrapper struct {
		Group Group `yaml:"group"`
	}
	bs, err := parseTemplate(path, tv, true)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(bs, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Group.Services, nil
}

func (c Squadron) Build(s Service) (string, error) {
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
	return Command(l, args...).Cwd(c.config.BasePath).Env(env).Run()
}

func (c Squadron) Push(name string) (string, error) {
	l := c.config.Log
	s, err := c.Service(name)
	if err != nil {
		return "", fmt.Errorf("could not find service: %w", err)
	}
	image := fmt.Sprintf("%s:%s", s.Image, s.Tag)

	l.Infof("Pushing service %v to %s", s.Name, image)

	return Command(l, "docker", "push", image).Cwd(c.config.BasePath).Run()
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
	// correct the relative path for the file:// chart repository
	wrapper.Service.Chart.Repository =
		strings.Replace(wrapper.Service.Chart.Repository, "file://./", fmt.Sprintf("file://%v/", basePath), 1)

	wrapper.Service.Chart.Alias = name
	return wrapper.Service, nil
}

func Init(l *logrus.Entry, dir string) (string, error) {
	l.Infof("Downloading example configuration into dir: %q", dir)
	return "", exampledata.RestoreAssets(dir, "")
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
	return stringInSlice(filepath.Ext(file), []string{".yml, .yaml"})
}

func isJson(file string) bool {
	return filepath.Ext(file) == ".json"
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

func parseTemplate(file string, templateVars interface{}, errOnMissing bool) ([]byte, error) {
	tmp, err := template.ParseFiles(file)
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer([]byte{})
	if errOnMissing {
		tmp = tmp.Option("missingkey=error")
	}
	if err := tmp.Funcs(builder.TemplateFuncs).Execute(out, templateVars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

type Cmd struct {
	l             *logrus.Entry
	cmd           *exec.Cmd
	wait          bool
	t             time.Duration
	preStartFunc  func() error
	postStartFunc func() error
	postEndFunc   func() error
	stdoutWriters []io.Writer
	stderrWriters []io.Writer
}

func Command(l *logrus.Entry, command ...string) *Cmd {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Env = os.Environ()
	return &Cmd{
		l:    l,
		cmd:  cmd,
		wait: true,
	}
}

func (c *Cmd) Cwd(path string) *Cmd {
	c.cmd.Dir = path
	return c
}

func (c *Cmd) Env(env []string) *Cmd {
	c.cmd.Env = append(c.cmd.Env, env...)
	return c
}

func (c *Cmd) Stdin(r io.Reader) *Cmd {
	c.cmd.Stdin = r
	return c
}

func (c *Cmd) Stdout(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stdoutWriters = append(c.stdoutWriters, w)
	return c
}

func (c *Cmd) Stderr(w io.Writer) *Cmd {
	if w == nil {
		w, _ = os.Open(os.DevNull)
	}
	c.stderrWriters = append(c.stderrWriters, w)
	return c
}

func (c *Cmd) Timeout(t time.Duration) *Cmd {
	c.t = t
	return c
}

func (c *Cmd) NoWait() *Cmd {
	c.wait = false
	return c
}

func (c *Cmd) PreStart(f func() error) *Cmd {
	c.preStartFunc = f
	return c
}

func (c *Cmd) PostStart(f func() error) *Cmd {
	c.postStartFunc = f
	return c
}

func (c *Cmd) PostEnd(f func() error) *Cmd {
	c.postEndFunc = f
	return c
}

func (c *Cmd) Run() (string, error) {
	c.l.Tracef("executing %q", c.cmd.String())

	combinedBuf := new(bytes.Buffer)
	traceWriter := c.l.WriterLevel(logrus.TraceLevel)
	warnWriter := c.l.WriterLevel(logrus.WarnLevel)

	c.stdoutWriters = append(c.stdoutWriters, combinedBuf, traceWriter)
	c.stderrWriters = append(c.stderrWriters, combinedBuf, warnWriter)
	c.cmd.Stdout = io.MultiWriter(c.stdoutWriters...)
	c.cmd.Stderr = io.MultiWriter(c.stderrWriters...)

	if c.preStartFunc != nil {
		if err := c.preStartFunc(); err != nil {
			return "", err
		}
	}

	if err := c.cmd.Start(); err != nil {
		return "", err
	}

	if c.postStartFunc != nil {
		if err := c.postStartFunc(); err != nil {
			return "", err
		}
	}

	if c.wait {
		if c.t != 0 {
			timer := time.AfterFunc(c.t, func() {
				c.cmd.Process.Kill()
			})
			defer timer.Stop()
		}

		if err := c.cmd.Wait(); err != nil {
			if c.t == 0 {
				return "", err
			}
		}
		if c.postEndFunc != nil {
			if err := c.postEndFunc(); err != nil {
				return "", err
			}
		}
	}

	return combinedBuf.String(), nil
}
