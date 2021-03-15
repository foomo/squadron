package squadron

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"text/template"

	"github.com/foomo/config-bob/builder"
	"github.com/foomo/squadron/util"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

const (
	defaultOutputDir    = ".output"
	chartApiVersionV2   = "v2"
	defaultChartType    = "application" // application or library
	chartFile           = "Chart.yaml"
	valuesFile          = "values.yaml"
	defaultSquadronFile = "squadron.yaml"
)

type Chart struct {
	APIVersion   string            `yaml:"apiVersion"`
	Name         string            `yaml:"name,omitempty"`
	Description  string            `yaml:"description,omitempty"`
	Type         string            `yaml:"type,omitempty"`
	Version      string            `yaml:"version,omitempty"`
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

func (c *Chart) addDependency(alias string, cd ChartDependency) {
	cd.Alias = alias
	c.Dependencies = append(c.Dependencies, cd)
}

func loadChart(path string) (*Chart, error) {
	c := Chart{}
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error while opening file: %v", err)
	}
	if err := yaml.Unmarshal(file, &c); err != nil {
		return nil, fmt.Errorf("error while unmarshalling template file: %s", err)
	}
	return &c, nil
}

func (c Chart) generate(chartPath string, overrides interface{}) error {
	// generate Chart.yaml
	if err := util.GenerateYaml(path.Join(chartPath, chartFile), c); err != nil {
		return err
	}
	// generate values.yaml
	if err := util.GenerateYaml(path.Join(chartPath, valuesFile), overrides); err != nil {
		return err
	}
	return nil
}

type Build struct {
	Image      string   `yaml:"image,omitempty"`
	Tag        string   `yaml:"tag,omitempty"`
	Context    string   `yaml:"context,omitempty"`
	Dockerfile string   `yaml:"dockerfile,omitempty"`
	Args       []string `yaml:"args,omitempty"`
	Labels     []string `yaml:"labels,omitempty"`
	CacheFrom  []string `yaml:"cache_from,omitempty"`
	Network    string   `yaml:"network,omitempty"`
	Target     string   `yaml:"target,omitempty"`
	ShmSize    string   `yaml:"shm_size,omitempty"`
	ExtraHosts []string `yaml:"extra_hosts,omitempty"`
	Isolation  string   `yaml:"isolation,omitempty"`
}

type ChartDependency struct {
	Name       string `yaml:"name,omitempty"`
	Repository string `yaml:"repository,omitempty"`
	Version    string `yaml:"version,omitempty"`
	Alias      string `yaml:"alias,omitempty"`
}

type Unit struct {
	Chart  interface{} `yaml:"chart,omitempty"`
	Build  interface{} `yaml:"build,omitempty"`
	Values interface{} `yaml:"values,omitempty"`
}

func (u Unit) GetChartDependency() (*ChartDependency, error) {
	cd := ChartDependency{}
	if vString, ok := u.Chart.(string); ok {
		localChart, err := loadChart(path.Join(vString, chartFile))
		if err != nil {
			return nil, fmt.Errorf("unable to load chart path %q, error: %v", vString, err)
		}
		cd.Name = localChart.Name
		cd.Repository = vString
		cd.Version = localChart.Version
	} else if vStruct, ok := u.Chart.(ChartDependency); ok {
		cd = vStruct
	} else {
		return nil, fmt.Errorf("incorrect format %q for chart field", u.Chart)
	}
	return &cd, nil
}

func (u Unit) GetBuild() (*Build, error) {
	b := Build{}
	if vString, ok := u.Build.(string); ok {
		b.Context = vString
	} else if vStruct, ok := u.Build.(Build); ok {
		b = vStruct
	} else {
		return nil, fmt.Errorf("invalid format %s for build field", u.Build)
	}
	return &b, nil
}

type Configuration struct {
	name    string
	Version string          `yaml:"version,omitempty"`
	Prefix  string          `yaml:"prefix,omitempty"`
	Units   map[string]Unit `yaml:"squadron,omitempty"`
}

type Squadron struct {
	l         *logrus.Entry
	helmCmd   *util.HelmCmd
	dockerCmd *util.DockerCmd
	basePath  string
	c         *Configuration
}

func New(l *logrus.Entry, basePath, namespace string) (*Squadron, error) {
	sq := Squadron{
		l:         l,
		helmCmd:   util.NewHelmCommand(l),
		dockerCmd: util.NewDockerCommand(l),
		basePath:  basePath,
	}
	sq.helmCmd.Args("-n", namespace)

	// execute without errors to get existing values
	out, err := executeFileTemplate(path.Join(basePath, defaultSquadronFile), nil, false)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(out, &sq.c); err != nil {
		return nil, err
	}

	// load existing values as template vars and execute again
	tv := sq.c.loadTemplateVars()
	out, err = executeFileTemplate(path.Join(basePath, defaultSquadronFile), tv, true)
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(out, &sq.c); err != nil {
		return nil, err
	}
	return &sq, nil
}

func (c Configuration) loadTemplateVars() TemplateVars {
	return TemplateVars{"Squadron": c.Units}
}

func (sq Squadron) Units() map[string]Unit {
	return sq.c.Units
}

func (sq Squadron) Up(units map[string]Unit, namespace string, helmArgs ...string) error {
	chartPath := path.Join(sq.basePath, defaultOutputDir, namespace, sq.c.name)
	// cleanup old files
	if err := sq.cleanupOutput(chartPath); err != nil {
		return err
	}
	// generate Chart.yaml and values.yaml
	if err := sq.generateChart(units, chartPath, sq.c.name, sq.c.Version); err != nil {
		return err
	}
	log.Println()
	// run helm dependancy upgrade
	_, err := sq.helmCmd.UpdateDependency(sq.c.name, chartPath)
	if err != nil {
		return err
	}
	// run helm upgrade --install --create-namespace
	sq.helmCmd.Args(helmArgs...)
	_, err = sq.helmCmd.Install(sq.c.name, chartPath)
	return err
}

func (sq Squadron) Build(u Unit) error {
	b, err := u.GetBuild()
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	dockerCmd := sq.dockerCmd
	dockerCmd.Option("-t", fmt.Sprintf("%v:%v", b.Image, b.Tag))
	dockerCmd.Option("--file", b.Dockerfile)
	dockerCmd.ListOption("--build-arg", b.Args)
	dockerCmd.ListOption("--label", b.Labels)
	dockerCmd.ListOption("--cache-from", b.CacheFrom)
	dockerCmd.Option("--network", b.Network)
	dockerCmd.Option("--target", b.Target)
	dockerCmd.Option("--shm-size", b.ShmSize)
	dockerCmd.ListOption("--add-host", b.ExtraHosts)
	dockerCmd.Option("--isolation", b.Isolation)
	_, err = sq.dockerCmd.Build(b.Context)
	return err
}

func (sq Squadron) Push(u Unit) error {
	b, err := u.GetBuild()
	if err != nil {
		return err
	}
	if b == nil {
		return nil
	}
	_, err = sq.dockerCmd.Push(b.Image, b.Tag)
	return err
}

func (sq Squadron) cleanupOutput(chartPath string) error {
	if _, err := os.Stat(chartPath); err == nil {
		sq.l.Infof("removing dir: %q", chartPath)
		if err := os.RemoveAll(chartPath); err != nil {
			sq.l.Warnf("could not delete chart output directory: %q", err)
		}
	}

	sq.l.Printf("creating dir: %q", chartPath)
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(chartPath, 0744); err != nil {
			return fmt.Errorf("could not create chart output directory: %w", err)
		}
	}
	return nil
}

func (sq Squadron) generateChart(units map[string]Unit, chartPath, chartName, version string) error {
	sq.l.Printf("generating chart %q files in %q", chartName, chartPath)
	chart := newChart(chartName, version)
	overrides := map[string]interface{}{}
	for name, unit := range units {
		cd, err := unit.GetChartDependency()
		if err != nil {
			return err
		}
		chart.addDependency(name, *cd)
		overrides[name] = unit.Values
	}
	if err := chart.generate(chartPath, overrides); err != nil {
		return err
	}
	return nil
}

type TemplateVars map[string]interface{}

func executeFileTemplate(path string, templateVars interface{}, errorOnMissing bool) ([]byte, error) {
	tplFuncs := builder.TemplateFuncs

	templateBytes, errRead := ioutil.ReadFile(path)
	if errRead != nil {
		return nil, errRead
	}
	tpl, err := template.New("squadron").Funcs(tplFuncs).Parse(string(templateBytes))
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer([]byte{})
	if errorOnMissing {
		tpl = tpl.Option("missingkey=error")
	}
	if err := tpl.Funcs(tplFuncs).Execute(out, templateVars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
