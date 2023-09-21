package squadron

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
	"text/template"

	"github.com/foomo/squadron/internal/config"
	templatex "github.com/foomo/squadron/internal/template"
	"github.com/foomo/squadron/internal/util"
	"github.com/miracl/conflate"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
	"github.com/sergi/go-diff/diffmatchpatch"
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
	var out bytes.Buffer
	t, err := template.New("namespace").Parse(sq.namespace)
	if err != nil {
		return "", err
	}
	if err := t.Execute(&out, map[string]string{"Squadron": squadron, "Unit": unit}); err != nil {
		return "", err
	}
	return out.String(), nil
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

func (sq *Squadron) MergeConfigFiles() error {
	pterm.Debug.Println("merging config files")
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
	if sq.c.Version != "2.0" {
		return errors.New("Please upgrade your YAML definition to 2.0")
	}

	sq.c.Trim()

	value, err := yamlv2.Marshal(sq.c)
	if err != nil {
		return err
	}

	sq.config = string(value)

	return nil
}

func (sq *Squadron) FilterConfig(squadron string, units, tags []string) error {
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
		if err := sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
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

	sq.c.Trim()

	value, err := yamlv2.Marshal(sq.c)
	if err != nil {
		return err
	}

	sq.config = string(value)

	return nil
}

func (sq *Squadron) RenderConfig(ctx context.Context) error {
	pterm.Debug.Println("rendering config")

	var tv templatex.Vars
	var vars map[string]interface{}
	if err := yaml.Unmarshal([]byte(sq.config), &vars); err != nil {
		return err
	}
	// execute again with loaded template vars
	tv = templatex.Vars{}
	if value, ok := vars["global"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Squadron", value)
	}

	// execute without errors to get existing values
	pterm.Debug.Println("executing file template")
	// pterm.Debug.Println(sq.config)
	out, err := templatex.ExecuteFileTemplate(ctx, sq.config, tv, false)
	if err != nil {
		return errors.Wrapf(err, "failed to execute initial file template\n%s", util.Highlight(sq.config))
	}

	// re-execute for rendering copied values
	pterm.Debug.Println("re-executing file template")
	// pterm.Debug.Println(string(out))
	out, err = templatex.ExecuteFileTemplate(ctx, string(out), tv, false)
	if err != nil {
		return errors.Wrap(err, "failed to re-execute initial file template")
	}

	pterm.Debug.Println("unmarshalling vars")
	if err := yaml.Unmarshal(out, &vars); err != nil {
		pterm.Error.Println(string(out))
		return errors.Wrap(err, "failed to unmarshal vars")
	}

	// execute again with loaded template vars
	tv = templatex.Vars{}
	if value, ok := vars["global"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Global", value)
	}
	if value, ok := vars["squadron"]; ok {
		util.ToSnakeCaseKeys(value)
		tv.Add("Squadron", value)
	}

	pterm.Debug.Println("executing file template")
	out, err = templatex.ExecuteFileTemplate(ctx, sq.config, tv, true)
	if err != nil {
		return errors.Wrap(err, "failed to execute second file template")
	}

	pterm.Debug.Println("unmarshalling vars")
	if err := yaml.Unmarshal(out, &sq.c); err != nil {
		pterm.Error.Println(string(out))
		return errors.Wrap(err, "failed to unmarshal vars")
	}

	sq.config = string(out)

	return nil
}

func (sq *Squadron) Push(ctx context.Context, pushArgs []string, parallel int) error {
	wg, gctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			wg.Go(func() error {
				if out, err := v.Push(gctx, key, k, pushArgs); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})
	})

	return wg.Wait()
}

func (sq *Squadron) Build(ctx context.Context, buildArgs []string, parallel int) error {
	wg, gctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			wg.Go(func() error {
				if out, err := v.Build(gctx, key, k, buildArgs); err != nil {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})
	})

	return wg.Wait()
}

func (sq *Squadron) Down(ctx context.Context, helmArgs []string, parallel int) error {
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			wg.Go(func() error {
				name := fmt.Sprintf("%s-%s", key, k)
				namespace, err := sq.Namespace(ctx, key, k)
				if err != nil {
					return err
				}

				stdErr := bytes.NewBuffer([]byte{})
				pterm.Debug.Printfln("running helm uninstall for: %s", name)
				if out, err := util.NewHelmCommand().Args("uninstall", name).
					Stderr(stdErr).
					Stdout(os.Stdout).
					Args("--namespace", namespace).
					Args(helmArgs...).
					Run(ctx); err != nil &&
					string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", name) {
					return errors.Wrap(err, out)
				}
				return nil
			})
			return nil
		})
	})

	return wg.Wait()
}

func (sq *Squadron) Diff(ctx context.Context, helmArgs []string, parallel int) error {
	var diff string
	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			wg.Go(func() error {
				name := fmt.Sprintf("%s-%s", key, k)
				namespace, err := sq.Namespace(ctx, key, k)
				if err != nil {
					return err
				}
				valueBytes, err := v.ValuesYAML(sq.c.Global)
				if err != nil {
					return err
				}

				pterm.Debug.Printfln("running helm diff for: %s", k)
				manifest, err := exec.CommandContext(ctx, "helm", "get", "manifest", name, "--namespace", sq.namespace).CombinedOutput()
				if err != nil && string(bytes.TrimSpace(manifest)) != errHelmReleaseNotFound {
					return errors.Wrap(err, string(manifest))
				}
				cmd := exec.CommandContext(ctx, "helm", "upgrade", name,
					"--install",
					"--namespace", namespace,
					"--set", fmt.Sprintf("squadron=%s", key),
					"--set", fmt.Sprintf("unit=%s", k),
					"--values", "-",
					"--dry-run",
				)
				cmd.Stdin = bytes.NewReader(valueBytes)
				if strings.Contains(v.Chart.Repository, "file://") {
					cmd.Args = append(cmd.Args, "/"+strings.TrimPrefix(v.Chart.Repository, "file://"))
				} else {
					cmd.Args = append(cmd.Args, v.Chart.Name, "--repo", v.Chart.Repository, "--version", v.Chart.Version)
				}
				cmd.Args = append(cmd.Args, helmArgs...)
				out, err := cmd.CombinedOutput()
				if err != nil {
					return errors.Wrap(err, string(out))
				}

				dmp := diffmatchpatch.New()
				diff += dmp.DiffPrettyText(dmp.DiffMain(string(manifest), string(out), false))
				return nil
			})
			return nil
		})
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	fmt.Println(diff)

	return nil
}

func (sq *Squadron) Status(ctx context.Context, helmArgs []string, parallel int) error {
	tbd := pterm.TableData{
		{"Name", "Revision", "Status", "Deployed by", "Commit", "Branch", "Last deployed", "Notes"},
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
		gitBranch  string `json:"-"`
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			var status statusType
			name := fmt.Sprintf("%s-%s", key, k)
			namespace, err := sq.Namespace(ctx, key, k)
			if err != nil {
				return err
			}

			stdErr := bytes.NewBuffer([]byte{})
			pterm.Debug.Printfln("running helm status for %s", name)
			if out, err := util.NewHelmCommand().Args("status", name).
				Stderr(stdErr).
				Args("--namespace", namespace, "--output", "json", "--show-desc").
				Args(helmArgs...).Run(ctx); err != nil && string(bytes.TrimSpace(stdErr.Bytes())) == errHelmReleaseNotFound {
				tbd = append(tbd, []string{name, "0", "not installed", "", ""})
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
					} else if strings.HasPrefix(line, "Git-Branch: ") {
						status.gitBranch = strings.TrimPrefix(line, "Git-Branch: ")
					} else {
						notes = append(notes, line)
					}
				}
				tbd = append(tbd, []string{
					status.Name,
					fmt.Sprintf("%d", status.Version),
					status.Info.Status,
					status.deployedBy,
					status.gitCommit,
					status.gitBranch,
					status.Info.LastDeployed, strings.Join(notes, " | "),
				})
			}
			return nil
		})
	})

	if err := wg.Wait(); err != nil {
		return err
	}

	return pterm.DefaultTable.WithHasHeader().WithData(tbd).Render()
}

func (sq *Squadron) Rollback(ctx context.Context, revision string, helmArgs []string, parallel int) error {
	if revision != "" {
		helmArgs = append([]string{revision}, helmArgs...)
	}

	wg, ctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			name := fmt.Sprintf("%s-%s", key, k)
			namespace, err := sq.Namespace(ctx, key, k)
			if err != nil {
				return err
			}

			stdErr := bytes.NewBuffer([]byte{})
			pterm.Debug.Printfln("running helm uninstall for: `%s`", name)
			if out, err := util.NewHelmCommand().Args("rollback", name).
				Stderr(stdErr).
				Stdout(os.Stdout).
				Args(helmArgs...).
				Args("--namespace", namespace).
				Run(ctx); err != nil &&
				string(bytes.TrimSpace(stdErr.Bytes())) != fmt.Sprintf("Error: uninstall: Release not loaded: %s: release: not found", name) {
				return errors.Wrap(err, out)
			}

			return nil
		})
	})

	return wg.Wait()
}

func (sq *Squadron) Up(ctx context.Context, helmArgs []string, username, version, commit, branch string, parallel int) error {
	description := fmt.Sprintf("\nDeployed-By: %s\nManaged-By: Squadron %s\nGit-Commit: %s\nGit-Branch: %s", username, version, commit, branch)

	wg, gctx := errgroup.WithContext(ctx)
	wg.SetLimit(parallel)

	_ = sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			wg.Go(func() error {
				name := fmt.Sprintf("%s-%s", key, k)
				namespace, err := sq.Namespace(ctx, key, k)
				if err != nil {
					return err
				}
				valueBytes, err := v.ValuesYAML(sq.c.Global)
				if err != nil {
					return err
				}

				// update local chart dependencies
				// https://stackoverflow.com/questions/59210148/error-found-in-chart-yaml-but-missing-in-charts-directory-mysql
				if strings.HasPrefix(v.Chart.Repository, "file:///") {
					pterm.Debug.Printfln("running helm dependency update for %s", v.Chart.Repository)
					if out, err := util.NewHelmCommand().
						Stdout(os.Stdout).
						Args("dependency", "update").
						Cwd(strings.TrimPrefix(v.Chart.Repository, "file://")).
						Run(ctx); err != nil {
						return errors.Wrap(err, out)
					}
				}

				// install chart
				pterm.Debug.Printfln("running helm upgrade for %s", name)
				cmd := util.NewHelmCommand().
					Stdin(bytes.NewReader(valueBytes)).
					Stdout(os.Stdout).
					Args("upgrade", name, "--install").
					Args("--set", fmt.Sprintf("squadron=%s,unit=%s", key, k)).
					Args("--description", description).
					Args("--namespace", namespace).
					Args("--dependency-update").
					Args("--install").
					Args("-values", "-").
					Args(helmArgs...)

				if strings.Contains(v.Chart.Repository, "file://") {
					cmd.Args(strings.TrimPrefix(v.Chart.Repository, "file://"))
				} else {
					cmd.Args(v.Chart.Name, "--repo", v.Chart.Repository, "--version", v.Chart.Version)
				}

				if out, err := cmd.Run(gctx); err != nil {
					return errors.Wrap(err, out)
				}

				return nil
			})
			return nil
		})
	})

	return wg.Wait()
}

func (sq *Squadron) Template(ctx context.Context, helmArgs []string) (string, error) {
	var ret bytes.Buffer

	err := sq.Config().Squadrons.Iterate(func(key string, value config.Map[*config.Unit]) error {
		return value.Iterate(func(k string, v *config.Unit) error {
			name := fmt.Sprintf("%s-%s", key, k)
			namespace, err := sq.Namespace(ctx, key, k)
			if err != nil {
				return err
			}
			valueBytes, err := v.ValuesYAML(sq.c.Global)
			if err != nil {
				return err
			}

			pterm.Debug.Printfln("running helm template for chart: %s", name)
			cmd := util.NewHelmCommand().Args("template", k).
				Stdin(bytes.NewReader(valueBytes)).
				Stdout(&ret).
				Args("--dependency-update").
				Args("--namespace", namespace).
				Args("--set", fmt.Sprintf("squadron=%s", key)).
				Args("--set", fmt.Sprintf("unit=%s", k)).
				Args("--values", "-").
				Args(helmArgs...)
			if strings.Contains(v.Chart.Repository, "file://") {
				file := strings.TrimPrefix(v.Chart.Repository, "file://")
				if !strings.HasPrefix(file, ".") {
					file = "/" + file
				}
				cmd.Args(file)
			} else {
				cmd.Args(v.Chart.Name, "--repo", v.Chart.Repository, "--version", v.Chart.Version)
			}
			if out, err := cmd.Run(ctx); err != nil {
				return errors.Wrap(err, out)
			}

			return nil
		})
	})
	if err != nil {
		return "", err
	}

	return ret.String(), nil
}
