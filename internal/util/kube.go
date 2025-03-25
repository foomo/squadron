package util

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	k8s "k8s.io/api/apps/v1"
)

type KubeCmd struct {
	Cmd
}

func NewKubeCommand() *KubeCmd {
	return &KubeCmd{*NewCommand("kubectl")}
}

func (c KubeCmd) RollbackDeployment(deployment string) *Cmd {
	return c.Args("rollout", "undo", fmt.Sprintf("deployment/%v", deployment))
}

func (c KubeCmd) WaitForRollout(deployment, timeout string) *Cmd {
	return c.Args("rollout", "status", fmt.Sprintf("deployment/%v", deployment),
		"-w", "--timeout", timeout)
}

func (c KubeCmd) GetMostRecentPodBySelectors(ctx context.Context, selectors map[string]string) (string, error) {
	var selector []string //nolint:prealloc
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	out, err := c.Args("--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime", "-o", "name").Run(ctx)
	if err != nil {
		return "", err
	}

	pods, err := parseResources(out, "pod/")
	if err != nil {
		return "", err
	}
	if len(pods) > 0 {
		return pods[len(pods)-1], nil
	}
	return "", errors.New("no pods found")
}

func (c KubeCmd) WaitForPodState(pod, condition, timeout string) *Cmd {
	return c.Args("wait", fmt.Sprintf("pod/%v", pod),
		fmt.Sprintf("--for=%v", condition),
		fmt.Sprintf("--timeout=%v", timeout))
}

func (c KubeCmd) ExecShell(resource, path string) *Cmd {
	return c.Args("exec", "-it", resource,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd %v && /bin/sh", path),
	).Stdin(os.Stdin).Stdout(os.Stdout).Stderr(os.Stdout)
}

func (c KubeCmd) PatchDeployment(patch, deployment string) *Cmd {
	return c.Args("patch", "deployment", deployment, "--patch", patch)
}

func (c KubeCmd) CopyToPod(pod, container, source, destination string) *Cmd {
	return c.Args("cp", source, fmt.Sprintf("%v:%v", pod, destination), "-c", container)
}

func (c KubeCmd) ExecPod(pod, container string, cmd []string) *Cmd {
	return c.Args("exec", pod, "-c", container, "--").Args(cmd...)
}

func (c KubeCmd) ExposePod(pod string, host string, port int) *Cmd {
	if host == "127.0.0.1" {
		host = ""
	}
	return c.Args("expose", "pod", pod, "--type=LoadBalancer",
		fmt.Sprintf("--port=%v", port), fmt.Sprintf("--external-ip=%v", host))
}

func (c KubeCmd) DeleteService(service string) *Cmd {
	return c.Args("delete", "service", service)
}

func (c KubeCmd) GetDeployment(ctx context.Context, deployment string) (*k8s.Deployment, error) {
	out, err := c.Args("get", "deployment", deployment, "-o", "json").Run(ctx)
	if err != nil {
		return nil, err
	}
	var d k8s.Deployment
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (c KubeCmd) GetNamespaces(ctx context.Context) ([]string, error) {
	out, err := c.Args("get", "namespace", "-o", "name").Run(ctx)
	if err != nil {
		return nil, err
	}
	return parseResources(out, "namespace/")
}

func (c KubeCmd) GetDeployments(ctx context.Context) ([]string, error) {
	out, err := c.Args("get", "deployment", "-o", "name").Run(ctx)
	if err != nil {
		return nil, err
	}
	return parseResources(out, "deployment.apps/")
}

func (c KubeCmd) GetPods(ctx context.Context, selectors map[string]string) ([]string, error) {
	var selector []string //nolint:prealloc
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	out, err := c.Args("--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name").Run(ctx)
	if err != nil {
		return nil, err
	}
	return parseResources(out, "pod/")
}

func (c KubeCmd) GetContainers(deployment k8s.Deployment) []string {
	containers := make([]string, len(deployment.Spec.Template.Spec.Containers))
	for i, c := range deployment.Spec.Template.Spec.Containers {
		containers[i] = c.Name
	}
	return containers
}

func (c KubeCmd) GetPodsByLabels(ctx context.Context, labels []string) ([]string, error) {
	out, err := c.Args("get", "pods", "-l", strings.Join(labels, ","), "-o", "name", "-A").Run(ctx)
	if err != nil {
		return nil, err
	}
	return parseResources(out, "pod/")
}

func (c KubeCmd) RestartDeployment(deployment string) *Cmd {
	return c.Args("rollout", "restart", fmt.Sprintf("deployment/%v", deployment))
}

func (c KubeCmd) CreateConfigMapFromFile(ctx context.Context, name, path string) (string, error) {
	return c.Args("create", "configmap", name, "--from-file", path).Run(ctx)
}

func (c KubeCmd) CreateConfigMap(ctx context.Context, name string, keyMap map[string]string) (string, error) {
	c.Args("create", "configmap", name)
	for key, value := range keyMap {
		c.Args(fmt.Sprintf("--from-literal=%v=%v", key, value))
	}
	return c.Run(ctx)
}

func (c KubeCmd) DeleteConfigMap(ctx context.Context, name string) (string, error) {
	return c.Args("delete", "configmap", name).Run(ctx)
}

func (c KubeCmd) GetConfigMapKey(ctx context.Context, name, key string) (string, error) {
	key = strings.ReplaceAll(key, ".", "\\.")
	// jsonpath map key is not very fond of dots
	out, err := c.Args("get", "configmap", name, "-o",
		fmt.Sprintf("jsonpath={.data.%v}", key)).Run(ctx)
	if err != nil {
		return out, err
	}
	if out == "" {
		return out, fmt.Errorf("no key %q found in ConfigMap %q", key, name)
	}
	return out, nil
}

func parseResources(out, prefix string) ([]string, error) {
	var res []string //nolint:prealloc
	if out == "" {
		return res, nil
	}
	lines := strings.Split(out, "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, fmt.Errorf("delimiter %q not found in %q", "\n", out)
	}
	for _, line := range lines {
		if line == "" {
			continue
		}
		unprefixed := strings.TrimPrefix(line, prefix)
		if unprefixed == line {
			return nil, fmt.Errorf("prefix %q not found in %q", prefix, line)
		}
		res = append(res, strings.TrimPrefix(line, prefix))
	}
	return res, nil
}
