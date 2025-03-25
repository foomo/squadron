package squadron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/foomo/squadron/internal/config"
	"github.com/foomo/squadron/internal/jsonschema"
	templatex "github.com/foomo/squadron/internal/template"
	"github.com/foomo/squadron/internal/util"
	"github.com/miracl/conflate"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/sters/yaml-diff/yamldiff"
	"golang.org/x/sync/errgroup"
	yamlv2 "gopkg.in/yaml.v2"
	"gopkg.in/yaml.v3"
)

const (
	errHelmReleaseNotFound = "Error: release: not found"
)

type Squadron struct {
	basePath  string
	namespace string
	files     []string
	config    string
	c         config.Config
}

type statusDescription struct {
	ManagedBy  string `json:"managedBy,omitempty"`
	DeployedBy string `json:"deployedBy,omitempty"`
	GitCommit  string `json:"gitCommit,omitempty"`
	GitBranch  string `json:"gitBranch,omitempty"`
}

func New(basePath, namespace string, files []string) *Squadron {
	return &Squadron{
		basePath:  basePath,
		namespace: namespace,
		files:     files,
		c:         config.Config{},
	}
}

// ------------------------------------------------------------------------------------------------
// ~ Getter
// ------------------------------------------------------------------------------------------------

func (sq *Squadron) Namespace(ctx context.Context, squadron, unit string) (string, error) {
	if sq.namespace == "" {
		return "default", nil
	}
	return util.RenderTemplateString(sq.namespace, map[string]string{"Squadron": squadron, "Unit": unit})
}

func (sq *Squadron) Config() config.Config {
	return sq.c
}

func (sq *Squadron) ConfigYAML() string {
	return sq.config
}

// ------------------------------------------------------------------------------------------------
// ~ Public methods
// ------------------------------------------------------------------------------------------------

func (sq *Squadron) MergeConfigFiles(ctx context.Context) error {
	pterm.Debug.Println(strings.Join(append([]string{"merging config files"}, sq.files...), "\n└ "))

	mergedFiles, err := conflate.FromFiles(sq.files...)
	if err != nil {
		return errors.Wrap(err, "failed to conflate files")
	}
	fileBytes, err := mergedFiles.MarshalYAML()
	if err != nil {
		return errors.Wrap(err, "failed to marshal yaml")
	}
	if err := yaml.Unmarshal(fileBytes, &sq.c); err != nil {
		pterm.Error.Println(string(fileBytes))
		return err
	}
	if sq.c.Version != config.Version {
		pterm.Debug.Println(string(fileBytes))
		return errors.New("Please upgrade your YAML definition to from '" + sq.c.Version + "' to '" + config.Version + "'")
	}

	sq.c.Trim(ctx)

	value, err := yamlv2.Marshal(sq.c)
	if err != nil {
		return err
	}

	sq.config = string(value)

	return nil
}

func (sq *Squadron) FilterConfig(ctx context.Context, squadron string, units, tags []string) error {
	if len(squadron) > 0 {
		if err := sq.Config().Squadrons.Filter(squadron); err != nil {
			return err
		}
	}

	if len(squadron) > 0 && len(units) > 0 {
		if err := sq.Config().Squadrons[squadron].Filter(units...); err != nil {
			return err
		}
	}

	if len(tags) > 0 {
		if err := sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
			return value.FilterFn(func(k string, v *config.Unit) bool {
				for _, tag := range tags {
					if strings.HasPrefix(tag, "-") {
						if slices.Contains(v.Tags, config.Tag(strings.TrimPrefix(tag, "-"))) {
							return false
						}
					} else if !slices.Contains(v.Tags, config.Tag(tag)) {
						return false
					}
				}
				return true
			})
		}); err != nil {
			return err
		}
	}

	sq.c.Trim(ctx)

	value, err := yamlv2.Marshal(sq.c)
	if err != nil {
		return err
	}

	sq.config = string(value)

	return nil
}

func (sq *Squadron) RenderConfig(ctx context.Context) error {
	var tv templatex.Vars
	var vars map[string]any
	if err := yaml.Unmarshal([]byte(sq.config), &vars); err != nil {
		return errors.Wrap(err, "failed to render config")
	}
	// execute again with loaded template vars
	tv = templatex.Vars{}
	if value, ok := vars["global"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Global", value)
	}
	if value, ok := vars["vars"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Vars", value)
	}
	if value, ok := vars["squadron"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Squadron", value)
	}

	out1, err := templatex.ExecuteFileTemplate(ctx, sq.config, tv, false)
	if err != nil {
		return errors.Wrapf(err, "failed to execute initial file template\n%s", util.Highlight(sq.config))
	}

	// re-execute for rendering copied values
	out2, err := templatex.ExecuteFileTemplate(ctx, string(out1), tv, false)
	if err != nil {
		fmt.Print(util.Highlight(string(out1)))
		return errors.Wrap(err, "failed to re-execute initial file template")
	}

	if err := yaml.Unmarshal(out2, &vars); err != nil {
		fmt.Print(util.Highlight(string(out2)))
		return errors.Wrap(err, "failed to unmarshal vars")
	}

	// execute again with loaded template vars
	tv = templatex.Vars{}
	if value, ok := vars["global"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Global", value)
	}
	if value, ok := vars["vars"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Vars", value)
	}
	if value, ok := vars["squadron"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Squadron", value)
	}

	out3, err := templatex.ExecuteFileTemplate(ctx, sq.config, tv, true)
	if err != nil {
		fmt.Print(util.Highlight(sq.config))
		return errors.Wrap(err, "failed to execute second file template")
	}

	if err := yaml.Unmarshal(out3, &sq.c); err != nil {
		fmt.Print(util.Highlight(string(out3)))
		return errors.Wrap(err, "failed to unmarshal vars")
	}

	sq.config = string(out3)

	return nil
}

func (sq *Squadron) Push(ctx context.Context, pushArgs []string, parallel int) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	type one struct {
		spinner  *util.PTermSpinner
		squadron string
		unit     string
		item     config.Build
	}
	var all []one

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			for _, name := range v.BuildNames() {
				build := v.Builds[name]
				spinner := printer.NewSpinner(fmt.Sprintf("🚚 | %s/%s.%s (%s:%s)", key, k, name, build.Image, build.Tag))
				all = append(all, one{
					spinner:  spinner,
					squadron: key,
					unit:     k,
					item:     build,
				})
				spinner.Start()
			}
			return nil
		})
	})

	for _, a := range all {
		wg.Go(func() error {
			ctx := a.spinner.Inject(ctx)
			if err := ctx.Err(); err != nil {
				a.spinner.Warning(err.Error())
				return err
			}

			if out, err := a.item.PushImage(ctx, a.squadron, a.unit, pushArgs); err != nil {
				a.spinner.Fail(out)
				return err
			}

			a.spinner.Success()
			return nil
		})
	}

	return wg.Wait()
}

func (sq *Squadron) BuildDependencies(ctx context.Context, buildArgs []string, parallel int) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	dependencies := sq.c.BuildDependencies(ctx)
	for name, build := range dependencies {
		wg.Go(func() error {
			spinner := printer.NewSpinner(fmt.Sprintf("📦 | %s (%s:%s)", name, build.Image, build.Tag))
			spinner.Start()

			ctx := spinner.Inject(ctx)
			if err := ctx.Err(); err != nil {
				spinner.Warning(err.Error())
				return err
			}

			if out, err := build.Build(ctx, "", "", buildArgs); err != nil {
				spinner.Fail(out)
				return err
			}

			spinner.Success()
			return nil
		})
	}

	return wg.Wait()
}

func (sq *Squadron) Build(ctx context.Context, buildArgs []string, parallel int) error {
	if err := sq.BuildDependencies(ctx, buildArgs, parallel); err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	type one struct {
		spinner  *util.PTermSpinner
		squadron string
		unit     string
		item     config.Build
	}
	var all []one

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			for _, name := range v.BuildNames() {
				build := v.Builds[name]
				spinner := printer.NewSpinner(fmt.Sprintf("📦 | %s/%s.%s (%s:%s)", key, k, name, build.Image, build.Tag))
				all = append(all, one{
					spinner:  spinner,
					squadron: key,
					unit:     k,
					item:     build,
				})
				spinner.Start()
			}
			return nil
		})
	})

	for _, a := range all {
		wg.Go(func() error {
			ctx := a.spinner.Inject(ctx)
			if err := ctx.Err(); err != nil {
				a.spinner.Warning(err.Error())
				return err
			}

			if out, err := a.item.Build(ctx, a.squadron, a.unit, buildArgs); err != nil {
				a.spinner.Fail(out)
				return err
			}

			a.spinner.Success()
			return nil
		})
	}

	return wg.Wait()
}

func (sq *Squadron) Down(ctx context.Context, helmArgs []string, parallel int) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			wg.Go(func() error {
				spinner := printer.NewSpinner(fmt.Sprintf("🗑️ | %s/%s", key, k))
				spinner.Start()

				ctx := spinner.Inject(ctx)
				if err := ctx.Err(); err != nil {
					spinner.Warning(err.Error())
					return err
				}

				name := fmt.Sprintf("%s-%s", key, k)
				namespace, err := sq.Namespace(ctx, key, k)
				if err != nil {
					return err
				}

				if out, err := util.NewHelmCommand().Args("uninstall", name).
					Args("--namespace", namespace).
					Args(helmArgs...).
					Run(ctx); err != nil &&
					strings.TrimSpace(out) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", name) {
					spinner.Fail(out)
					return err
				}

				spinner.Success()
				return nil
			})
			return nil
		})
	})

	return wg.Wait()
}

func (sq *Squadron) RenderSchema(ctx context.Context, baseSchema string) (string, error) {
	js := jsonschema.New()
	if err := js.LoadBaseSchema(ctx, baseSchema); err != nil {
		return "", errors.Wrap(err, "failed to load base schema")
	}

	if err := sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			if err := ctx.Err(); err != nil {
				return err
			}

			if v.Chart.Schema == "" {
				return nil
			}

			return js.SetSquadronUnitSchema(ctx, key, k, v.Chart.Schema)
		})
	}); err != nil {
		return "", err
	}

	return js.PrettyString()
}

func (sq *Squadron) Diff(ctx context.Context, helmArgs []string, parallel int) (string, error) {
	var m sync.Mutex
	var ret bytes.Buffer
	write := func(b []byte) error {
		m.Lock()
		defer m.Unlock()
		_, err := ret.Write(b)
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			wg.Go(func() error {
				spinner := printer.NewSpinner(fmt.Sprintf("🔍 | %s/%s", key, k))
				spinner.Start()

				ctx := spinner.Inject(ctx)
				if err := ctx.Err(); err != nil {
					spinner.Warning(err.Error())
					return err
				}

				name := fmt.Sprintf("%s-%s", key, k)
				namespace, err := sq.Namespace(ctx, key, k)
				if err != nil {
					return err
				}
				valueBytes, err := v.ValuesYAML(sq.c.Global)
				if err != nil {
					return err
				}

				manifest, err := exec.CommandContext(ctx, "helm", "get", "manifest", name, "--namespace", namespace).CombinedOutput()
				if err != nil && string(bytes.TrimSpace(manifest)) != errHelmReleaseNotFound {
					spinner.Fail(string(manifest))
					return err
				}
				cmd := exec.CommandContext(ctx, "helm", "upgrade", name,
					"--install",
					"--namespace", namespace,
					"--set", "global.foomo.squadron.name="+key,
					"--set", "global.foomo.squadron.unit="+k,
					"--hide-notes",
					"--values", "-",
					"--dry-run",
				)
				cmd.Args = append(cmd.Args, v.PostRendererArgs()...)
				cmd.Stdin = bytes.NewReader(valueBytes)

				if strings.HasPrefix(v.Chart.Repository, "file://") {
					cmd.Args = append(cmd.Args, path.Clean(strings.TrimPrefix(v.Chart.Repository, "file://")))
				} else {
					cmd.Args = append(cmd.Args, v.Chart.Name)
					if v.Chart.Repository != "" {
						cmd.Args = append(cmd.Args, "--repo", v.Chart.Repository)
					}
					if v.Chart.Version != "" {
						cmd.Args = append(cmd.Args, "--version", v.Chart.Version)
					}
				}
				cmd.Args = append(cmd.Args, helmArgs...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					spinner.Fail(string(out))
					return err
				}

				yamls1, err := yamldiff.Load(string(manifest))
				if err != nil {
					spinner.Fail(string(manifest))
					return errors.Wrap(err, "failed to load yaml diff")
				}

				outStr := strings.Split(string(out), "\n")
				yamls2, err := yamldiff.Load(strings.Join(outStr[10:], "\n"))
				if err != nil {
					spinner.Fail(string(out))
					return errors.Wrap(err, "failed to load yaml diff")
				}

				var res string
				for _, diff := range yamldiff.Do(yamls1, yamls2) {
					res += diff.Dump() + "  ---\n"
				}

				if err := write([]byte(res)); err != nil {
					spinner.Fail(res)
					return err
				}

				spinner.Success()
				return nil
			})

			return nil
		})
	})

	if err := wg.Wait(); err != nil {
		return "", err
	}

	return ret.String(), nil
}

func (sq *Squadron) Status(ctx context.Context, helmArgs []string, parallel int) error {
	var m sync.Mutex
	tbd := pterm.TableData{
		{"Name", "Revision", "Status", "Managed by", "Deployed by", "Commit", "Branch", "Last deployed", "Notes"},
	}
	write := func(b []string) {
		m.Lock()
		defer m.Unlock()
		tbd = append(tbd, b)
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
		managedBy  string `json:"-"` //nolint:revive
		deployedBy string `json:"-"` //nolint:revive
		gitCommit  string `json:"-"` //nolint:revive
		gitBranch  string `json:"-"` //nolint:revive
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			var status statusType
			name := fmt.Sprintf("%s-%s", key, k)
			namespace, err := sq.Namespace(ctx, key, k)
			if err != nil {
				return errors.Errorf("failed to retrieve namsspace: %s/%s", key, k)
			}

			wg.Go(func() error {
				spinner := printer.NewSpinner(fmt.Sprintf("📄 | %s/%s", key, k))
				spinner.Start()

				ctx := spinner.Inject(ctx)
				if err := ctx.Err(); err != nil {
					spinner.Warning(err.Error())
					return err
				}

				stdErr := bytes.NewBuffer([]byte{})
				out, err := util.NewHelmCommand().Args("status", name).
					Stderr(stdErr).
					Args("--namespace", namespace, "--output", "json", "--show-desc").
					Args(helmArgs...).Run(ctx)

				if err != nil && string(bytes.TrimSpace(stdErr.Bytes())) == errHelmReleaseNotFound {
					tbd = append(tbd, []string{name, "0", "not installed", "", ""})
				} else if err != nil {
					spinner.Fail(stdErr.String())
					return err
				}

				if err := json.Unmarshal([]byte(out), &status); err != nil {
					spinner.Fail(out)
					return errors.Errorf("failed to retrieve status: %s/%s", key, k)
				}

				var notes []string
				lines := strings.Split(status.Info.Description, "\n")
				var statusDescription statusDescription

				if err := json.Unmarshal([]byte(status.Info.Description), &statusDescription); err == nil {
					status.managedBy = statusDescription.ManagedBy
					status.deployedBy = statusDescription.DeployedBy
					status.gitCommit = statusDescription.GitCommit
					status.gitBranch = statusDescription.GitBranch
				} else if len(lines) > 1 {
					for _, line := range lines {
						if strings.HasPrefix(line, "Managed-By: ") {
							status.managedBy = strings.TrimPrefix(line, "Managed-By: ")
						} else if strings.HasPrefix(line, "Deployed-By: ") {
							status.deployedBy = strings.TrimPrefix(line, "Deployed-By: ")
						} else if strings.HasPrefix(line, "Git-Commit: ") {
							status.gitCommit = strings.TrimPrefix(line, "Git-Commit: ")
						} else if strings.HasPrefix(line, "Git-Branch: ") {
							status.gitBranch = strings.TrimPrefix(line, "Git-Branch: ")
						} else {
							notes = append(notes, line)
						}
					}
				} else {
					notes = append(notes, status.Info.Description)
				}

				write([]string{
					status.Name,
					fmt.Sprintf("%d", status.Version),
					status.Info.Status,
					status.managedBy,
					status.deployedBy,
					status.gitCommit,
					status.gitBranch,
					status.Info.LastDeployed,
					strings.Join(notes, "\n"),
				})

				spinner.Success()
				return nil
			})

			return nil
		})
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	out, err := pterm.DefaultTable.WithHasHeader().WithData(tbd).Srender()
	if err != nil {
		return err
	}
	pterm.Println(out)

	return nil
}

func (sq *Squadron) Rollback(ctx context.Context, revision string, helmArgs []string, parallel int) error {
	if revision != "" {
		helmArgs = append([]string{revision}, helmArgs...)
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			name := fmt.Sprintf("%s-%s", key, k)
			namespace, err := sq.Namespace(ctx, key, k)
			if err != nil {
				return err
			}

			wg.Go(func() error {
				spinner := printer.NewSpinner(fmt.Sprintf("♻️ | %s/%s", key, k))
				spinner.Start()

				ctx := spinner.Inject(ctx)
				if err := ctx.Err(); err != nil {
					spinner.Warning(err.Error())
					return err
				}

				stdErr := bytes.NewBuffer([]byte{})
				out, err := util.NewHelmCommand().Args("rollback", name).
					Stderr(stdErr).
					// Stdout(os.Stdout).
					Args(helmArgs...).
					Args("--namespace", namespace).
					Run(ctx)
				if err != nil &&
					string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", name) {
					spinner.Fail(stdErr.String())
					return err
				}

				spinner.Success(out)
				return nil
			})

			return nil
		})
	})

	return wg.Wait()
}

// UpdateLocalDependencies work around
// https://stackoverflow.com/questions/59210148/error-found-in-chart-yaml-but-missing-in-charts-directory-mysql
func (sq *Squadron) UpdateLocalDependencies(ctx context.Context, parallel int) error {
	// collect unique entrie
	repositories := map[string]struct{}{}
	err := sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			if strings.HasPrefix(v.Chart.Repository, "file:///") {
				repositories[v.Chart.Repository] = struct{}{}
			}
			return nil
		})
	})
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	for repository := range repositories {
		wg.Go(func() error {
			pterm.Debug.Printfln("running helm dependency update for %s", repository)
			if out, err := util.NewHelmCommand().
				Cwd(path.Clean(strings.TrimPrefix(repository, "file://"))).
				Args("dependency", "update", "--skip-refresh", "--debug").
				Run(ctx); err != nil {
				return errors.Wrap(err, out)
			} else {
				pterm.Debug.Println(out)
			}
			return nil
		})
	}

	return wg.Wait()
}

func (sq *Squadron) Up(ctx context.Context, helmArgs []string, username, version, commit, branch string, parallel int) error {
	description, err := json.Marshal(statusDescription{
		ManagedBy:  version,
		DeployedBy: username,
		GitCommit:  commit,
		GitBranch:  branch,
	})
	if err != nil {
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()
	type one struct {
		spinner  *util.PTermSpinner
		squadron string
		unit     string
		item     *config.Unit
	}
	var all []one

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			spinner := printer.NewSpinner(fmt.Sprintf("🚀 | %s/%s", key, k))
			all = append(all, one{
				spinner:  spinner,
				squadron: key,
				unit:     k,
				item:     v,
			})
			spinner.Start()

			return nil
		})
	})

	for _, a := range all {
		wg.Go(func() error {
			ctx := a.spinner.Inject(ctx)
			if err := ctx.Err(); err != nil {
				a.spinner.Warning(err.Error())
				return err
			}

			name := fmt.Sprintf("%s-%s", a.squadron, a.unit)
			namespace, err := sq.Namespace(ctx, a.squadron, a.unit)
			if err != nil {
				a.spinner.Fail(err.Error())
				return err
			}
			valueBytes, err := a.item.ValuesYAML(sq.c.Global)
			if err != nil {
				a.spinner.Fail(err.Error())
				return err
			}

			// install chart
			cmd := util.NewHelmCommand().
				Stdin(bytes.NewReader(valueBytes)).
				// Stdout(os.Stdout).
				Args("upgrade", name, "--install").
				Args("--set", "global.foomo.squadron.name="+a.squadron).
				Args("--set", "global.foomo.squadron.unit="+a.unit).
				Args("--description", string(description)).
				Args("--namespace", namespace).
				Args("--dependency-update").
				Args(a.item.PostRendererArgs()...).
				Args("--install").
				Args("--values", "-").
				Args(helmArgs...)

			if strings.HasPrefix(a.item.Chart.Repository, "file://") {
				cmd.Args(path.Clean(strings.TrimPrefix(a.item.Chart.Repository, "file://")))
			} else {
				cmd.Args(a.item.Chart.Name)
				if a.item.Chart.Repository != "" {
					cmd.Args("--repo", a.item.Chart.Repository)
				}
				if a.item.Chart.Version != "" {
					cmd.Args("--version", a.item.Chart.Version)
				}
			}

			out, err := cmd.Run(ctx)
			if err != nil {
				a.spinner.Fail(out)
				return err
			}

			a.spinner.Success()
			return nil
		})
	}

	return wg.Wait()
}

func (sq *Squadron) Template(ctx context.Context, helmArgs []string, parallel int) (string, error) {
	var m sync.Mutex
	var ret bytes.Buffer
	write := func(b []byte) error {
		m.Lock()
		defer m.Unlock()
		_, err := ret.Write(b)
		return err
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	printer := util.MustNewPTermMultiPrinter()
	defer printer.Stop()

	_ = sq.Config().Squadrons.Iterate(ctx, func(ctx context.Context, key string, value config.Map[*config.Unit]) error {
		return value.Iterate(ctx, func(ctx context.Context, k string, v *config.Unit) error {
			wg.Go(func() error {
				spinner := printer.NewSpinner(fmt.Sprintf("🧾 | %s/%s", key, k))
				spinner.Start()

				ctx := spinner.Inject(ctx)
				if err := ctx.Err(); err != nil {
					spinner.Warning(err.Error())
					return err
				}

				name := fmt.Sprintf("%s-%s", key, k)
				namespace, err := sq.Namespace(ctx, key, k)
				if err != nil {
					spinner.Fail(err.Error())
					return errors.Errorf("failed to retrieve namsspace: %s/%s", key, k)
				}

				out, err := v.Template(ctx, name, key, k, namespace, sq.c.Global, helmArgs)
				if err != nil {
					spinner.Fail(string(out))
					return err
				}

				if err := write(out); err != nil {
					spinner.Fail(string(out))
					return err
				}

				spinner.Success()
				return nil
			})

			return nil
		})
	})

	if err := wg.Wait(); err != nil {
		return "", err
	}

	return ret.String(), nil
}
