package squadron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/miracl/conflate"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/sirupsen/logrus"
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"

	"github.com/foomo/squadron/runner"

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

func (sq *Squadron) Name() string {
	return sq.name
}

func (sq *Squadron) GetConfig() Configuration {
	return sq.c
}

func (sq *Squadron) GetConfigYAML() string {
	return sq.config
}

func (sq *Squadron) MergeConfigFiles() error {
	logrus.Info("merging config files")
	pterm.Debug.Println(strings.Join(append([]string{"using files"}, sq.files...), "\n└ "))

	mergedFiles, err := conflate.FromFiles(sq.files...)
	if err != nil {
		return errors.Wrap(err, "failed to conflate files")
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
	logrus.Info("rendering config")
	var tv TemplateVars
	var vars map[string]interface{}
	if err := yaml.Unmarshal([]byte(sq.config), &vars); err != nil {
		return err
	}
	// execute again with loaded template vars
	tv = TemplateVars{}
	if value, ok := vars["global"]; ok {
		toSnakeCaseKeys(value)
		tv.add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		toSnakeCaseKeys(value)
		tv.add("Squadron", value)
	}

	// execute without errors to get existing values
	pterm.Debug.Println("executing file template")
	// pterm.Debug.Println(sq.config)
	out, err := executeFileTemplate(ctx, sq.config, tv, false)
	if err != nil {
		return errors.Wrap(err, "failed to execute initial file template")
	}

	// re-execute for rendering copied values
	pterm.Debug.Println("re-executing file template")
	// pterm.Debug.Println(string(out))
	out, err = executeFileTemplate(ctx, string(out), tv, false)
	if err != nil {
		return errors.Wrap(err, "failed to re-execute initial file template")
	}

	pterm.Debug.Println("unmarshalling vars")
	if err := yaml.Unmarshal(out, &vars); err != nil {
		pterm.Error.Println(string(out))
		return errors.Wrap(err, "failed to unmarshal vars")
	}

	// execute again with loaded template vars
	tv = TemplateVars{}
	if value, ok := vars["global"]; ok {
		toSnakeCaseKeys(value)
		tv.add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		toSnakeCaseKeys(value)
		tv.add("Squadron", value)
	}

	pterm.Debug.Println("executing file template")
	out, err = executeFileTemplate(ctx, sq.config, tv, true)
	if err != nil {
		return errors.Wrap(err, "failed to execute second file template")
	}
	pterm.Debug.Println("unmarshalling vars")
	if err := yaml.Unmarshal(out, &sq.c); err != nil {
		pterm.Error.Println(string(out))
		return errors.Wrap(err, "failed to unmarshal vars")
	}
	sq.config = string(out)

	if sq.c.Name != "" {
		sq.name = sq.c.Name
	}

	return nil
}

func (sq *Squadron) Generate(ctx context.Context, units Units) error {
	logrus.WithField("path", sq.chartPath()).Infof("generating charts")
	if err := sq.cleanupOutput(sq.chartPath()); err != nil {
		return err
	}
	if sq.c.Unite {
		return sq.generateUmbrellaChart(ctx, units)
	}
	for _, uName := range units.Keys() {
		u := units[uName]
		// update local chart dependencies
		// https://stackoverflow.com/questions/59210148/error-found-in-chart-yaml-but-missing-in-charts-directory-mysql
		if strings.HasPrefix(u.Chart.Repository, "file:///") {
			pterm.Debug.Printfln("running helm dependency update for %s", u.Chart.Repository)
			if out, err := util.NewHelmCommand().
				Stdout(os.Stdout).
				Args("dependency", "update").
				Cwd(strings.TrimPrefix(u.Chart.Repository, "file://")).
				Run(ctx); err != nil {
				return errors.Wrap(err, out)
			}
		}

		pterm.Debug.Printfln("generating %q value overrides file in %q", uName, sq.chartPath())
		if err := sq.generateValues(u.Values, sq.chartPath(), uName); err != nil {
			return err
		}
	}
	return nil
}

func (sq *Squadron) generateUmbrellaChart(ctx context.Context, units Units) error {
	pterm.Debug.Printfln("generating chart %q files in %q", sq.name, sq.chartPath())
	if err := sq.generateChart(units, sq.chartPath(), sq.name, sq.c.Version); err != nil {
		return err
	}
	pterm.Debug.Printfln("running helm dependency update for chart: %v", sq.chartPath())
	if out, err := util.NewHelmCommand().UpdateDependency(ctx, sq.chartPath()); err != nil {
		return errors.Wrap(err, out)
	}
	return nil
}

func (sq *Squadron) Package(ctx context.Context) error {
	pterm.Debug.Printfln("running helm package for chart: %v", sq.chartPath())
	if out, err := util.NewHelmCommand().Package(ctx, sq.chartPath(), sq.basePath); err != nil {
		return errors.Wrap(err, out)
	}
	return nil
}

func (sq *Squadron) Down(ctx context.Context, units Units, helmArgs []string) error {
	if sq.c.Unite {
		pterm.Debug.Printfln("running helm uninstall for: %s", sq.chartPath())
		stdErr := bytes.NewBuffer([]byte{})
		if out, err := util.NewHelmCommand().Args("uninstall", sq.name).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", sq.name) {
			return errors.Wrap(err, out)
		}
	}
	for _, uName := range units.Keys() {
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		pterm.Debug.Printfln("running helm uninstall for: %s", uName)
		stdErr := bytes.NewBuffer([]byte{})
		if out, err := util.NewHelmCommand().Args("uninstall", rName).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", rName) {
			return errors.Wrap(err, out)
		}
	}

	return nil
}

func (sq *Squadron) Diff(ctx context.Context, units Units, helmArgs []string) (string, error) {
	if sq.c.Unite {
		pterm.Debug.Printfln("running helm diff for: %s", sq.chartPath())
		manifest, err := exec.CommandContext(ctx, "helm", "get", "manifest", sq.name, "--namespace", sq.namespace).CombinedOutput() //nolint:gosec
		if err != nil {
			return "", errors.Wrap(err, string(manifest))
		}
		cmd := exec.CommandContext(ctx, "helm", "upgrade", sq.name, sq.chartPath(), "--namespace", sq.namespace, "--dry-run") //nolint:gosec
		cmd.Args = append(cmd.Args, helmArgs...)
		template, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrap(err, string(template))
		}
		dmp := diffmatchpatch.New()
		return dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(template), false)), nil
	}
	for _, uName := range units.Keys() {
		u := units[uName]
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		pterm.Debug.Printfln("running helm diff for: %s", uName)
		manifest, err := exec.CommandContext(ctx, "helm", "get", "manifest", rName, "--namespace", sq.namespace).CombinedOutput()
		if err != nil && string(bytes.TrimSpace(manifest)) != errHelmReleaseNotFound {
			return "", errors.Wrap(err, string(manifest))
		}
		cmd := exec.CommandContext(ctx, "helm", "upgrade", rName,
			"--install",
			"--namespace", sq.namespace,
			"-f", path.Join(sq.chartPath(), uName+".yaml"),
			"--set", fmt.Sprintf("squadron=%s,unit=%s", sq.name, uName),
			"--dry-run",
		)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args = append(cmd.Args, "/"+strings.TrimPrefix(u.Chart.Repository, "file://"))
		} else {
			cmd.Args = append(cmd.Args, u.Chart.Name, "--repo", u.Chart.Repository, "--version", u.Chart.Version)
		}
		cmd.Args = append(cmd.Args, helmArgs...)
		template, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrap(err, string(template))
		}
		dmp := diffmatchpatch.New()
		_, _ = fmt.Println(dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(template), false)))
	}

	return "", nil
}

func (sq *Squadron) Status(ctx context.Context, units Units, helmArgs []string) error {
	tbd := pterm.TableData{
		{"Name", "Revision", "Status", "Deployed by", "Commit", "Last deployed", "Notes"},
	}

	type statusType struct {
		Name      string `json:"name"`
		Version   int    `json:"version"`
		Namespace string `json:"namespace"`
		Info      struct {
			Status        string `json:"status"`
			FirstDeployed string `json:"first_deployed"`
			Deleted       string `json:"deleted"`
			LastDeployed  string `json:"last_deployed"`
			Description   string `json:"description"`
		} `json:"info"`
		deployedBy string `json:"-"`
		gitCommit  string `json:"-"`
	}

	var status statusType

	if sq.c.Unite {
		stdErr := bytes.NewBuffer([]byte{})
		pterm.Debug.Printfln("running helm status for chart: %s", sq.chartPath())
		if out, err := util.NewHelmCommand().Args("status", sq.name).
			Stderr(stdErr).
			Args("--namespace", sq.namespace, "--output", "json", "--show-desc").
			Args(helmArgs...).
			Run(ctx); err != nil && string(bytes.TrimSpace(stdErr.Bytes())) == errHelmReleaseNotFound {
			tbd = append(tbd, []string{sq.name, "0", "not installed", "", ""})
		} else if err != nil {
			return errors.Wrap(err, out)
		} else if err := json.Unmarshal([]byte(out), &status); err != nil {
			return errors.Wrap(err, out)
		} else {
			var notes []string
			for _, line := range strings.Split(status.Info.Description, "\n") {
				if strings.HasPrefix(line, "Managed-By: ") {
					// do nothing
				} else if strings.HasPrefix(line, "Deployed-By: ") {
					status.deployedBy = strings.TrimPrefix(line, "Deployed-By: ")
				} else if strings.HasPrefix(line, "Git-Commit: ") {
					status.gitCommit = strings.TrimPrefix(line, "Git-Commit: ")
				} else {
					notes = append(notes, line)
				}
			}
			tbd = append(tbd, []string{status.Name, fmt.Sprintf("%d", status.Version), status.Info.Status, status.deployedBy, status.gitCommit, status.Info.LastDeployed, strings.Join(notes, " | ")})
		}
	}
	for _, uName := range units.Keys() {
		stdErr := bytes.NewBuffer([]byte{})
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		pterm.Debug.Printfln("running helm status for %s", uName)
		if out, err := util.NewHelmCommand().Args("status", rName).
			Stderr(stdErr).
			Args("--namespace", sq.namespace, "--output", "json", "--show-desc").
			Args(helmArgs...).Run(ctx); err != nil && string(bytes.TrimSpace(stdErr.Bytes())) == errHelmReleaseNotFound {
			tbd = append(tbd, []string{rName, "0", "not installed", "", ""})
		} else if err != nil {
			return errors.Wrap(err, out)
		} else if err := json.Unmarshal([]byte(out), &status); err != nil {
			return errors.Wrap(err, out)
		} else {
			var notes []string
			for _, line := range strings.Split(status.Info.Description, "\n") {
				if strings.HasPrefix(line, "Managed-By: ") {
					// do nothing
				} else if strings.HasPrefix(line, "Deployed-By: ") {
					status.deployedBy = strings.TrimPrefix(line, "Deployed-By: ")
				} else if strings.HasPrefix(line, "Git-Commit: ") {
					status.gitCommit = strings.TrimPrefix(line, "Git-Commit: ")
				} else {
					notes = append(notes, line)
				}
			}
			tbd = append(tbd, []string{status.Name, fmt.Sprintf("%d", status.Version), status.Info.Status, status.deployedBy, status.gitCommit, status.Info.LastDeployed, strings.Join(notes, " | ")})
		}
	}

	return pterm.DefaultTable.WithHasHeader().WithData(tbd).Render()
}

func (sq *Squadron) Rollback(ctx context.Context, units Units, revision string, helmArgs []string) error {
	if revision != "" {
		helmArgs = append([]string{revision}, helmArgs...)
	}
	if sq.c.Unite {
		pterm.Debug.Printfln("running helm rollback for: %s", sq.chartPath())
		stdErr := bytes.NewBuffer([]byte{})
		if out, err := util.NewHelmCommand().Args("rollback", sq.name).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args(helmArgs...).
			Args("--namespace", sq.namespace).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", sq.name) {
			return errors.Wrap(err, out)
		}
	}
	for _, uName := range units.Keys() {
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		pterm.Debug.Printfln("running helm uninstall for: %s", uName)
		stdErr := bytes.NewBuffer([]byte{})
		if out, err := util.NewHelmCommand().Args("rollback", rName).
			Stderr(stdErr).
			Stdout(os.Stdout).
			Args(helmArgs...).
			Args("--namespace", sq.namespace).
			Run(ctx); err != nil &&
			string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", rName) {
			return errors.Wrap(err, out)
		}
	}

	return nil
}

func (sq *Squadron) Up(ctx context.Context, units Units, helmArgs []string, username, version, commit string, parallel int) error {
	description := fmt.Sprintf("\nDeployed-By: %s\nManaged-By: Squadron %s\nGit-Commit: %s", username, version, commit)

	if sq.c.Unite {
		pterm.Debug.Printfln("running helm upgrade for chart: %s", sq.chartPath())
		if out, err := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("upgrade", sq.name, sq.chartPath()).
			Args("--namespace", sq.namespace).
			Args("--dependency-update").
			Args("--description", description).
			Args("--install").
			Args(helmArgs...).
			Run(ctx); err != nil {
			return errors.Wrap(err, out)
		}
		return nil
	}
	r := runner.Runner{}

	for _, uName := range units.Keys() {
		u := units[uName]
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)

		// update local chart dependencies
		// https://stackoverflow.com/questions/59210148/error-found-in-chart-yaml-but-missing-in-charts-directory-mysql
		if strings.HasPrefix(u.Chart.Repository, "file:///") {
			pterm.Debug.Printfln("running helm dependency update for %s", u.Chart.Repository)
			if out, err := util.NewHelmCommand().
				Stdout(os.Stdout).
				Args("dependency", "update").
				Cwd(strings.TrimPrefix(u.Chart.Repository, "file://")).
				Run(ctx); err != nil {
				return errors.Wrap(err, out)
			}
		}

		// install chart
		pterm.Debug.Printfln("running helm upgrade for %s", uName)
		cmd := util.NewHelmCommand().
			Stdout(os.Stdout).
			Args("upgrade", rName, "--install").
			Args("--set", fmt.Sprintf("squadron=%s,unit=%s", sq.name, uName)).
			Args("--description", description).
			Args("--namespace", sq.namespace).
			Args("--dependency-update").
			Args("--install").
			Args("-f", path.Join(sq.chartPath(), uName+".yaml")).
			Args(helmArgs...)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args(strings.TrimPrefix(u.Chart.Repository, "file://"))
		} else {
			cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository, "--version", u.Chart.Version)
		}

		r.Add(func(tctx context.Context) error {
			if out, err := cmd.Run(tctx); err != nil {
				return errors.Wrap(err, out)
			}
			return nil
		})
	}
	return r.Run(ctx, parallel)
}

func (sq *Squadron) Template(ctx context.Context, units Units, helmArgs []string) error {
	if sq.c.Unite {
		pterm.Debug.Printfln("running helm template for chart: %s", sq.chartPath())
		if out, err := util.NewHelmCommand().Args("template", sq.name, sq.chartPath()).
			Stdout(os.Stdout).
			Args("--dependency-update").
			Args("--namespace", sq.namespace).
			Args(helmArgs...).
			Run(ctx); err != nil {
			return errors.Wrap(err, out)
		}
		return nil
	}
	for _, uName := range units.Keys() {
		u := units[uName]
		// todo use release prefix on install: squadron name or --name
		rName := fmt.Sprintf("%s-%s", sq.name, uName)
		pterm.Debug.Printfln("running helm template for chart: %s", uName)
		cmd := util.NewHelmCommand().Args("template", rName).
			Stdout(os.Stdout).
			Args("--dependency-update").
			Args("--namespace", sq.namespace).
			Args("--set", fmt.Sprintf("squadron=%s,unit=%s", sq.name, uName)).
			Args("-f", path.Join(sq.chartPath(), uName+".yaml")).
			Args(helmArgs...)
		if strings.Contains(u.Chart.Repository, "file://") {
			cmd.Args("/" + strings.TrimPrefix(u.Chart.Repository, "file://"))
		} else {
			cmd.Args(u.Chart.Name, "--repo", u.Chart.Repository, "--version", u.Chart.Version)
		}
		if out, err := cmd.Run(ctx); err != nil {
			return errors.Wrap(err, out)
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

func (sq *Squadron) generateChart(units Units, chartPath, chartName, version string) error {
	chart := newChart(chartName, version)
	values := map[string]interface{}{}
	if sq.c.Global != nil {
		values["global"] = sq.c.Global
	}
	_ = units.Iterate(func(name string, unit *Unit) error {
		chart.addDependency(name, unit.Chart)
		values[name] = unit.Values
		return nil
	})
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
