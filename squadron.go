package squadron

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/miracl/conflate"
	"github.com/pkg/errors"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/util"
)

func init() {
	yamlv2.FutureLineWrap()
}

const (
	defaultOutputDir       = ".squadron"
	chartAPIVersionV2      = "v2"
	defaultChartType       = "application" // application or library
	chartFile              = "Chart.yaml"
	valuesFile             = "values.yaml"
	errHelmReleaseNotFound = "Error: release: not found"
)

type Configuration struct {
	Name    string                 `yaml:"name,omitempty"`
	Version string                 `yaml:"version,omitempty"`
	Prefix  string                 `yaml:"prefix,omitempty"`
	Unite   bool                   `yaml:"unite,omitempty"`
	Global  map[string]interface{} `yaml:"global,omitempty"`
	Units   map[string]*Unit       `yaml:"squadron,omitempty"`
}

type Squadron struct {
	name      string
	basePath  string
	namespace string
	files     []string
	config    string
	c         Configuration
}

func New(basePath, namespace string, files []string) *Squadron {
	return &Squadron{
		name:      filepath.Base(basePath),
		basePath:  basePath,
		namespace: namespace,
		files:     files,
		c:         Configuration{},
	}
}

func (sq *Squadron) GetConfig() Configuration {
	return sq.c
}

func (sq *Squadron) GetConfigYAML() string {
	return sq.config
}

// UnmarshalYAML ...
func (c *Configuration) UnmarshalYAML(value *yaml.Node) error {
	if value.Tag == TagMap {
		type wrapper Configuration
		err := value.Decode((*wrapper)(c))
		if err == nil {
			// if the decode is successful, remove units that are nil
			c.removeNilUnits()
		}
		return err
	}
	return fmt.Errorf("unsupported node tag type for %T: %q", c, value.Tag)
}

func (c *Configuration) removeNilUnits() {
	for uName, u := range c.Units {
		if u == nil {
			delete(c.Units, uName)
		}
	}
}

func (sq *Squadron) MergeConfigFiles() error {
	mergedFiles, err := conflate.FromFiles(sq.files...)
	if err != nil {
		return errors.Wrap(err, "failed to conflate files")
	}
	var data interface{}
	if err := mergedFiles.Unmarshal(&data); err != nil {
		return errors.Wrap(err, "failed to unmarshal data")
	}
	fileBytes, err := mergedFiles.MarshalYAML()
	if err != nil {
		return errors.Wrap(err, "failed to marshal yaml")
	}
	if err := yaml.Unmarshal(fileBytes, &sq.c); err != nil {
		return err
	}
	sq.config = string(fileBytes)
	return nil
}

func (sq *Squadron) FilterConfig(units []string) error {
	unitsMap := make(map[string]bool, len(units))
	for _, unit := range units {
		unitsMap[unit] = true
	}

	for name := range sq.c.Units {
		if _, ok := unitsMap[name]; !ok {
			delete(sq.c.Units, name)
		}
	}
	value, err := yaml.Marshal(sq.c)
	if err != nil {
		return err
	}
	sq.config = string(value)
	return nil
}

func (sq *Squadron) RenderConfig(ctx context.Context) error {
	var tv TemplateVars
	var vars map[string]interface{}
	if err := yaml.Unmarshal([]byte(sq.config), &vars); err != nil {
		return err
	}
	// execute again with loaded template vars
	tv = TemplateVars{}
	if value, ok := vars["global"]; ok {
		replace(value)
		tv.add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		replace(value)
		tv.add("Squadron", value)
	}
	// execute without errors to get existing values
	out, err := executeFileTemplate(ctx, sq.config, tv, false)
	if err != nil {
		return errors.Wrap(err, "failed to execute initial file template")
	}

	if err := yaml.Unmarshal(out, &vars); err != nil {
		return err
	}
	// execute again with loaded template vars
	tv = TemplateVars{}
	if value, ok := vars["global"]; ok {
		replace(value)
		tv.add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		replace(value)
		tv.add("Squadron", value)
	}
	out, err = executeFileTemplate(ctx, sq.config, tv, true)
	if err != nil {
		return errors.Wrap(err, "failed to execute second file template")
	}
	if err := yaml.Unmarshal(out, &sq.c); err != nil {
		return err
	}
	sq.config = string(out)

	if sq.c.Name != "" {
		sq.name = sq.c.Name
	}

	return nil
}

func (sq *Squadron) Generate(ctx context.Context, units map[string]*Unit) error {
	logrus.Infof("recreating chart output dir %q", sq.chartPath())
	if err := sq.cleanupOutput(sq.chartPath()); err != nil {
		return err
	}
	if sq.c.Unite {
		return sq.generateUmbrellaChart(ctx, units)
	}
	for uName, u := range units {
		logrus.Infof("generating %q value overrides file in %q", uName, sq.chartPath())
		if err := sq.generateValues(u.Values, sq.chartPath(), uName); err != nil {
			return err
		}
	}
	return nil
}

func (sq *Squadron) generateUmbrellaChart(ctx context.Context, units map[string]*Unit) error {
	logrus.Infof("generating chart %q files in %q", sq.name, sq.chartPath())
	if err := sq.generateChart(units, sq.chartPath(), sq.name, sq.c.Version); err != nil {
		return err
	}
	logrus.Infof("running helm dependency update for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().UpdateDependency(ctx, sq.chartPath())
	return err
}

func (sq *Squadron) Package(ctx context.Context) error {
	logrus.Infof("running helm package for chart: %v", sq.chartPath())
	_, err := util.NewHelmCommand().Package(ctx, sq.chartPath(), sq.basePath)
	return err
}

func (sq *Squadron) Down(ctx context.Context, units map[string]*Unit, helmArgs []string) error {
	if sq.c.Unite {
		logrus.Infof("running helm uninstall for: %s", sq.chartPath())
		stdErr := bytes.NewBuffer([]byte{})
		if _, err := util.NewHelmCommand().Args("uninstall", sq.name).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", sq.name) {
			return err
		}
	}
	for uName := range units {
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm uninstall for: %s", uName)
		stdErr := bytes.NewBuffer([]byte{})
		if _, err := util.NewHelmCommand().Args("uninstall", rName).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", rName) {
			return err
		}
	}

	return nil
}

func (sq *Squadron) Diff(ctx context.Context, units map[string]*Unit, helmArgs []string) (string, error) {
	if sq.c.Unite {
		logrus.Infof("running helm diff for: %s", sq.chartPath())
		manifest, err := exec.CommandContext(ctx, "helm", "get", "manifest", sq.name, "--namespace", sq.namespace).Output() //nolint:gosec
		if err != nil {
			return "", err
		}
		template, err := exec.CommandContext(ctx, "helm", "upgrade", sq.name, sq.chartPath(), "--namespace", sq.namespace, "--dry-run").Output() //nolint:gosec
		if err != nil {
			return "", err
		}
		dmp := diffmatchpatch.New()
		return dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(template), false)), nil
	}
	for uName, u := range units {
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm diff for: %s", uName)
		manifest, err := exec.CommandContext(ctx, "helm", "get", "manifest", rName, "--namespace", sq.namespace).CombinedOutput()
		if err != nil && string(bytes.TrimSpace(manifest)) != errHelmReleaseNotFound {
			return "", err
		}
		cmd := exec.CommandContext(ctx, "helm", "upgrade", rName, "--install", "--namespace", sq.namespace, "-f", path.Join(sq.chartPath(), uName+".yaml"), "--dry-run")
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args = append(cmd.Args, "/"+strings.TrimPrefix(u.Chart.Repository, "file://"))
		} else {
			cmd.Args = append(cmd.Args, u.Chart.Name, "--repo", u.Chart.Repository)
		}
		template, err := cmd.Output()
		if err != nil {
			return "", err
		}
		dmp := diffmatchpatch.New()
		return dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(template), false)), nil
	}

	return "", nil
}

func (sq *Squadron) Status(ctx context.Context, units map[string]*Unit, helmArgs []string) error {
	stdOut := bytes.NewBuffer([]byte{})
	if sq.c.Unite {
		stdOut.WriteString("==== " + sq.name + strings.Repeat("=", 20-len(sq.name)) + "\n")
		logrus.Infof("running helm status for chart: %s", sq.chartPath())
		stdErr := bytes.NewBuffer([]byte{})
		if _, err := util.NewHelmCommand().Args("status", sq.name).
			Stderr(stdErr).
			Stdout(stdOut).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) == errHelmReleaseNotFound {
			stdOut.WriteString("NAME: " + sq.name + "\n")
			stdOut.WriteString("STATUS: not installed\n")
		} else if err != nil {
			return err
		}
	}
	for uName := range units {
		stdOut.WriteString("==== " + uName + " " + strings.Repeat("=", 60-len(uName)) + "\n")
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm status for %s", uName)
		stdErr := bytes.NewBuffer([]byte{})
		if _, err := util.NewHelmCommand().Args("status", rName).
			Stderr(stdErr).
			Stdout(stdOut).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) == errHelmReleaseNotFound {
			stdOut.WriteString("NAME: " + rName + "\n")
			stdOut.WriteString("STATUS: not installed\n")
		} else if err != nil {
			return err
		}
	}
	fmt.Println(strings.ReplaceAll(stdOut.String(), "\\n", "\n"))
	return nil
}

func (sq *Squadron) Up(ctx context.Context, units map[string]*Unit, helmArgs []string, username, version, commit string) error {
	description := fmt.Sprintf("\nDeployed-By: %s\nManaged-By: Squadron %s\nGit-Commit: %s", version, username, commit)

	if sq.c.Unite {
		logrus.Infof("running helm upgrade for chart: %s", sq.chartPath())
		_, err := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("upgrade", sq.name, sq.chartPath()).
			Args("--namespace", sq.namespace).
			Args("--dependency-update").
			Args("--description", description).
			Args("--install").
			Args(helmArgs...).
			Run(ctx)
		return err
	}
	for uName, u := range units {
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		// logrus.Infof(
		// 	"running helm dependency update for %s in %s",
		// 	uName,
		// 	strings.TrimPrefix(u.Chart.Repository, "file://"),
		// )
		// if strings.Contains(u.Chart.Repository, "file://") {
		// 	if _, err := util.NewHelmCommand().
		// 		Args("dependency", "update").
		// 		Cwd(strings.TrimPrefix(u.Chart.Repository, "file://")).
		// 		Stdout(os.Stdout).
		// 		Run(ctx); err != nil {
		// 		return err
		// 	}
		// }
		logrus.Infof("running helm upgrade for %s", uName)
		cmd := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("upgrade", rName, "--install").
			Args("--description", description).
			Args("--namespace", sq.namespace).
			Args("--dependency-update").
			Args("--install").
			Args("-f", path.Join(sq.chartPath(), uName+".yaml")).
			Args(helmArgs...)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args(strings.TrimPrefix(u.Chart.Repository, "file://"))
		} else {
			cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository)
		}
		if _, err := cmd.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (sq *Squadron) Template(ctx context.Context, units map[string]*Unit, helmArgs []string) error {
	if sq.c.Unite {
		logrus.Infof("running helm template for chart: %s", sq.chartPath())
		_, err := util.NewHelmCommand().Args("template", sq.name, sq.chartPath()).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx)
		return err
	}
	for uName, u := range units {
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		logrus.Infof("running helm template for chart: %s", uName)
		cmd := util.NewHelmCommand().Args("template", rName).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args("-f", path.Join(sq.chartPath(), uName+".yaml")).
			Args(helmArgs...)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args("/" + strings.TrimPrefix(u.Chart.Repository, "file://"))
		} else {
			cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository)
		}
		if _, err := cmd.Run(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (sq *Squadron) chartPath() string {
	return path.Join(sq.basePath, defaultOutputDir, sq.name)
}

func (sq *Squadron) cleanupOutput(chartPath string) error {
	if _, err := os.Stat(chartPath); err == nil {
		if err := os.RemoveAll(chartPath); err != nil {
			logrus.Warnf("could not delete chart output directory: %q", err)
		}
	}
	if _, err := os.Stat(chartPath); os.IsNotExist(err) {
		if err := os.MkdirAll(chartPath, 0o744); err != nil {
			return fmt.Errorf("could not create chart output directory: %w", err)
		}
	}
	return nil
}

func (sq *Squadron) generateChart(units map[string]*Unit, chartPath, chartName, version string) error {
	chart := newChart(chartName, version)
	values := map[string]interface{}{}
	if sq.c.Global != nil {
		values["global"] = sq.c.Global
	}
	for name, unit := range units {
		chart.addDependency(name, unit.Chart)
		values[name] = unit.Values
	}
	return chart.generate(chartPath, values)
}

func (sq *Squadron) generateValues(values map[string]interface{}, vPath, vName string) error {
	if values == nil {
		values = map[string]interface{}{}
	}
	if sq.c.Global != nil {
		values["global"] = sq.c.Global
	}
	return util.GenerateYaml(path.Join(vPath, vName+".yaml"), values)
}
