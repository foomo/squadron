package squadron

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/foomo/squadron/exampledata"
	"github.com/foomo/squadron/util"
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
	requirementsFile       = "requirements.yaml"
	chartApiVersionV2      = "v2"
	chartApiVersionV1      = "v1"
	defaultChartType       = "application" // application or library
	defaultChartVersion    = "0.2.0"
	defaultChartAppVersion = "1.16.0"
	defaultHelmRepo        = "https://kubernetes-charts.storage.googleapis.com/"
)

var (
	ErrServiceNotFound    = errors.New("service not found")
	ErrBuildNotConfigured = errors.New("build command was not configured")
)

type Override map[string]interface{}

type Group struct {
	Name             string `yaml:"-"`
	Version          string
	ServiceOverrides map[string]Override `yaml:"services"`
	JobOverrides     map[string]Override `yaml:"jobs"`
}

func (g Group) Services() []string {
	var services []string
	for service := range g.ServiceOverrides {
		services = append(services, service)
	}
	return services
}

type Namespace struct {
	name   string
	groups []string
}

type Config struct {
	Tag      string
	BasePath string
	Log      *logrus.Entry
}

type Build struct {
	Image   string
	Tag     string
	Command string
}

type Service struct {
	Name  string          `yaml:"-"`
	Build Build           `yaml:"build"`
	Chart ChartDependency `yaml:"chart"`
}
type ChartDependency struct {
	Name       string
	Repository string
	Version    string
	Alias      string
}

func (cd *ChartDependency) validate(basePath, service string) error {
	if cd.Name == "" {
		return fmt.Errorf("service %q chart field %q required", service, "name")
	}
	if cd.Version == "" {
		return fmt.Errorf("service %q chart field %q required", service, "version")
	}
	if strings.HasPrefix(cd.Repository, "file://./") {
		cd.Repository = strings.Replace(cd.Repository, "file://./", fmt.Sprintf("file://%v/", basePath), 1)
	}
	if cd.Repository == "" {
		cd.Repository = defaultHelmRepo
	}
	cd.Alias = service
	return nil
}

type Chart struct {
	APIVersion   string `yaml:"apiVersion"`
	Name         string
	Description  string
	Type         string
	Version      string
	Dependencies []ChartDependency `yaml:"dependencies,omitempty"`
}

func newChart(name, version string) *Chart {
	return &Chart{
		APIVersion:  chartApiVersionV2,
		Name:        name,
		Description: fmt.Sprintf("A helm parent chart for squadron %v", name),
		Type:        defaultChartType,
		Version:     version,
	}
}

func (c Chart) generateChartFiles(chartPath string, overrides interface{}, useChartApiV1 bool) error {
	if useChartApiV1 {
		c.APIVersion = chartApiVersionV1
		// for chart APIVersion v1 dependencies are defined outside Chart.yaml, in requirements.yaml
		var wrapper struct {
			Dependencies []ChartDependency
		}
		wrapper.Dependencies = c.Dependencies
		c.Dependencies = nil
		// generate requirements.yaml
		if err := util.GenerateYaml(path.Join(chartPath, requirementsFile), wrapper); err != nil {
			return err
		}
	}
	// generate Chart.yaml
	if err := util.GenerateYaml(path.Join(chartPath, chartFile), c); err != nil {
		return err
	}
	// generate values.yaml with overrides
	if err := util.GenerateYaml(path.Join(chartPath, valuesFile), overrides); err != nil {
		return err
	}
	return nil
}

type JobItem struct {
	Name      string
	Overrides interface{}
	namespace string
	group     string
	chart     string
}

type serviceLoader func(string) (Service, error)

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
	if !util.IsYaml(source) && !util.IsJson(source) {
		return fmt.Errorf("Unable to parse %q, supported: %v", source, strings.Join(tv.supportedFileExt(), ", "))
	}
	file, err := ioutil.ReadFile(source)
	if err != nil {
		return fmt.Errorf("Error while opening template file: %s", err)
	}
	if util.IsYaml(source) {
		if err := yaml.Unmarshal(file, &tv); err != nil {
			return fmt.Errorf("Error while unmarshalling template file: %s", err)
		}
	}
	if util.IsJson(source) {
		if err := json.Unmarshal(file, &tv); err != nil {
			return fmt.Errorf("Error while unmarshalling template file: %s", err)
		}
		return nil
	}
	return nil
}

type Squadron struct {
	l          *logrus.Entry
	basePath   string
	tag        string
	Services   []Service
	Templates  []string
	Namespaces []Namespace
	helmCmd    *util.HelmCmd
	kubeCmd    *util.KubeCmd
	dockerCmd  *util.DockerCmd
}

func New(l *logrus.Entry, tag, basePath, namespace string) (*Squadron, error) {
	sq := Squadron{l: l, basePath: basePath, tag: tag}
	sq.helmCmd = util.NewHelmCommand(l)
	sq.kubeCmd = util.NewKubeCommand(l)
	sq.dockerCmd = util.NewDockerCommand(l)
	sq.helmCmd.Args("-n", namespace)
	sq.kubeCmd.Args("-n", namespace)

	l.Infof("Parsing configuration files")
	l.Infof("Entering dir: %q", basePath)
	serviceDir := path.Join(basePath, defaultServiceDir)
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
			l.Infof("Loading service: %v, from: %q", name, util.RelativePath(path, basePath))
			svc, err := loadService(file, name, tag, basePath)
			if err != nil {
				return err
			}
			sq.Services = append(sq.Services, svc)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sq.Namespaces, err = loadNamespaces(l, sq.Service, basePath)
	if err != nil {
		return nil, err
	}

	return &sq, nil
}

func (sq Squadron) Service(name string) (Service, error) {
	var available []string
	for _, s := range sq.Services {
		if s.Name == name {
			return s, nil
		}
		available = append(available, s.Name)
	}
	return Service{}, errResourceNotFound(name, "service", available)
}

func (sq Squadron) Namespace(name string) (Namespace, error) {
	var available []string
	for _, ns := range sq.Namespaces {
		if ns.name == name {
			return ns, nil
		}
		available = append(available, ns.name)
	}
	return Namespace{}, errResourceNotFound(name, "namespace", available)
}

func (ns Namespace) ValidateGroup(name string) error {
	var available []string
	for _, g := range ns.groups {
		if g == name {
			return nil
		}
		available = append(available, g)
	}
	return errResourceNotFound(name, "squadron", available)
}

func (sq Squadron) Group(namespace, group string, tv TemplateVars) (Group, error) {
	var wrapper struct {
		Group Group `yaml:"squadron"`
	}
	ns, err := sq.Namespace(namespace)
	if err != nil {
		return wrapper.Group, err
	}
	if err := ns.ValidateGroup(group); err != nil {
		return wrapper.Group, err
	}

	path := path.Join(sq.basePath, defaultNamespaceDir, namespace, group+defaultConfigFileExt)
	bs, err := util.ExecuteTemplate(path, tv)
	if err != nil {
		return wrapper.Group, err
	}
	wrapper.Group.Name = group
	if err := yaml.Unmarshal(bs, &wrapper); err != nil {
		return wrapper.Group, err
	}
	return wrapper.Group, nil
}

func (sq Squadron) Build(s Service) (string, error) {
	sq.l.Infof("Building service: %v", s.Name)
	if s.Build.Command == "" {
		return "", ErrBuildNotConfigured
	}

	args := strings.Split(s.Build.Command, " ")
	if args[0] == "docker" && s.Build.Image != "" && s.Build.Tag != "" {
		image := fmt.Sprintf("%v:%v", s.Build.Image, s.Build.Tag)
		args = append(strings.Split(s.Build.Command, " "), "-t", image)
	}
	return util.NewCommand(sq.l, args[0]).Args(args[1:]...).Cwd(sq.basePath).Run()
}

func (sq Squadron) Push(s Service) (string, error) {
	image := fmt.Sprintf("%v:%v", s.Build.Image, s.Build.Tag)
	if s.Build.Image == "" || s.Build.Tag == "" {
		return "", fmt.Errorf("invalid image %q to build service %q", image, s.Name)
	}
	sq.l.Infof("Pushing service %v to %s", s.Name, image)
	return sq.dockerCmd.Push(s.Build.Image, s.Build.Tag)
}

func Init(l *logrus.Entry, dir string) (string, error) {
	l.Infof("Copying example configuration into dir: %q", dir)
	return "", exampledata.RestoreAssets(dir, "")
}

func (sq Squadron) CheckIngressController(name string) error {
	pods, err := sq.kubeCmd.GetPodsByLabels([]string{fmt.Sprintf("app.kubernetes.io/name=%v", name)})
	if err != nil {
		return fmt.Errorf("error while checking for ingress controller %q: %s", name, err)
	}
	if len(pods) == 0 {
		return fmt.Errorf("ingress controller %q not present on any namespace", name)
	}
	return nil
}

func (sq Squadron) Install(group Group, tv TemplateVars, outputDir string) (string, error) {
	groupChartPath := path.Join(sq.basePath, defaultOutputDir, outputDir, group.Name)
	if err := sq.prepareChart(group, tv, groupChartPath, false); err != nil {
		return "", err
	}
	return sq.helmCmd.Install(group.Name, groupChartPath)
}

func (sq Squadron) Generate(group Group, tv TemplateVars, outputDir string, useChartApiV1 bool) (string, error) {
	groupChartPath := path.Join(sq.basePath, defaultOutputDir, outputDir, group.Name)
	if err := sq.prepareChart(group, tv, groupChartPath, useChartApiV1); err != nil {
		return "", err
	}
	return sq.helmCmd.Package(group.Name, groupChartPath, path.Join(sq.basePath, defaultOutputDir, outputDir))
}

func (sq Squadron) Uninstall(group string) (string, error) {
	output, err := sq.helmCmd.Uninstall(group)
	if err != nil {
		return output, err
	}
	return output, nil
}

func (sq Squadron) Restart(services []string) (string, error) {
	for _, service := range services {
		sq.l.Infof("Waiting for service %q to get ready", service)
		out, err := sq.kubeCmd.WaitForRollout(service, "30s").Run()
		if err != nil {
			return out, err
		}
		sq.l.Infof("Restarting service %q", service)
		out, err = sq.kubeCmd.RestartDeployment(service).Run()
		if err != nil {
			return out, err
		}
	}
	return "", nil
}

func (sq Squadron) prepareChart(group Group, tv TemplateVars, groupChartPath string, useChartApiV1 bool) error {
	sq.l.Infof("Preparing chart")
	sq.l.Infof("Entering dir: %q", path.Join(sq.basePath, defaultOutputDir))
	if _, err := os.Stat(groupChartPath); err == nil {
		sq.l.Infof("Removing dir: %q", groupChartPath)
		if err := os.RemoveAll(groupChartPath); err != nil {
			sq.l.Warnf("could not delete group chart directory: %q", err)
		}
	}

	sq.l.Printf("Creating dir: %q", groupChartPath)
	if _, err := os.Stat(groupChartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(groupChartPath, 0744); err != nil {
			return fmt.Errorf("could not create a workdir directory: %w", err)
		}
	}

	groupChart := newChart(group.Name, group.Version)
	for _, service := range group.Services() {
		s, err := sq.Service(service)
		if err != nil {
			return err
		}
		groupChart.Dependencies = append(groupChart.Dependencies, s.Chart)
	}
	if err := groupChart.generateChartFiles(groupChartPath, group.ServiceOverrides, useChartApiV1); err != nil {
		return err
	}

	_, err := sq.helmCmd.UpdateDependency(group.Name, groupChartPath)
	if err != nil {
		return err
	}

	return nil
}

func loadService(reader io.Reader, name, tag, basePath string) (Service, error) {
	var wrapper struct {
		Service Service `yaml:"service"`
	}
	if err := yaml.NewDecoder(reader).Decode(&wrapper); err != nil {
		return Service{}, fmt.Errorf("could not decode service: %w", err)
	}
	wrapper.Service.Name = name
	if wrapper.Service.Build.Tag == "" {
		wrapper.Service.Build.Tag = tag
	}
	if err := wrapper.Service.Chart.validate(basePath, name); err != nil {
		return Service{}, err
	}
	return wrapper.Service, nil
}

func loadNamespaces(l *logrus.Entry, sl serviceLoader, basePath string) ([]Namespace, error) {
	var nss []Namespace
	namespaceDir := path.Join(basePath, defaultNamespaceDir)
	err := filepath.Walk(namespaceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != namespaceDir {
			l.Infof("Loading namespace: %v, from: %q", info.Name(), util.RelativePath(path, basePath))
			gs, err := loadGroupNames(l, sl, basePath, info.Name())
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

func loadGroupNames(l *logrus.Entry, sl serviceLoader, basePath, namespace string) ([]string, error) {
	var gs []string
	groupPath := path.Join(basePath, defaultNamespaceDir, namespace)
	err := filepath.Walk(groupPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, defaultConfigFileExt)) {
			name := strings.TrimSuffix(info.Name(), defaultConfigFileExt)
			l.Infof("Loading squadron: %v, from: %q", name, util.RelativePath(path, basePath))
			gs = append(gs, name)
		}
		return nil
	})
	return gs, err
}

func errResourceNotFound(name, resource string, available []string) error {
	if name == "" {
		return fmt.Errorf("%s not provided. Available: %s", resource, strings.Join(available, ", "))
	}
	return fmt.Errorf("%s '%s' not found. Available: %s", resource, name, strings.Join(available, ", "))
}
