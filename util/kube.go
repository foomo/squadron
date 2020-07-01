package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

type KubeCommand struct {
	*CliCommand
	l         *logrus.Entry
	namespace string
}

func NewKubeCommand(l *logrus.Entry, namespace string) *KubeCommand {
	return &KubeCommand{&CliCommand{"kubectl"}, l, namespace}
}

func (kc KubeCommand) RollbackDeployment(deployment string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"rollout", "undo", fmt.Sprintf("deployment/%v", deployment),
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) WaitForRollout(deployment, timeout string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"rollout", "status", fmt.Sprintf("deployment/%v", deployment),
		"-w", "--timeout", timeout,
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) GetMostRecentPodBySelectors(selectors map[string]string) (string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name",
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return "", err
	}

	pods, err := parseResources(out, "\n", "pod/")
	if err != nil {
		return "", err
	}
	if len(pods) > 0 {
		return pods[len(pods)-1], nil
	}
	return "", fmt.Errorf("no pods found")
}

func (kc KubeCommand) WaitForPodState(pod, condition, timeout string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"wait", fmt.Sprintf("pod/%v", pod),
		fmt.Sprintf("--for=%v", condition),
		fmt.Sprintf("--timeout=%v", timeout),
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) ExecShell(resource, path string) *Cmd {
	cmdArgs := []string{
		kc.name, "-n", kc.namespace,
		"exec", "-it", resource,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd %v && /bin/sh", path),
	}

	return Command(kc.l, cmdArgs...).Stdin(os.Stdin).Stdout(os.Stdout).Stderr(os.Stdout)
}

func (kc KubeCommand) PatchDeployment(patch, deployment string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"patch", "deployment", deployment,
		"--patch", patch,
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) CopyToPod(pod, container, source, destination string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"cp", source, fmt.Sprintf("%v:%v", pod, destination),
		"-c", container,
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) ExecPod(pod, container string, cmd []string) *Cmd {
	c := []string{
		kc.name, "-n", kc.namespace,
		"exec", pod,
		"-c", container,
		"--",
	}
	c = append(c, cmd...)
	return Command(kc.l, c...)
}

func (kc KubeCommand) ExposePod(pod string, host string, port int) *Cmd {
	if host == "127.0.0.1" {
		host = ""
	}
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"expose", "pod", pod,
		"--type=LoadBalancer",
		fmt.Sprintf("--port=%v", port),
		fmt.Sprintf("--external-ip=%v", host),
		// fmt.Sprintf("--name=%v-%v", pod, port),
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) DeleteService(service string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"delete", "service", service,
	}
	return Command(kc.l, cmd...)
}

func (kc KubeCommand) GetDeployment(deployment string) (*v1.Deployment, error) {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"get", "deployment", deployment,
		"-o", "json",
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}
	var d v1.Deployment
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (kc KubeCommand) GetNamespaces() ([]string, error) {
	cmd := []string{
		kc.name,
		"get", "namespace",
		"-o", "name",
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "namespace/")
}

func (kc KubeCommand) GetDeployments() ([]string, error) {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"get", "deployment",
		"-o", "name",
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "deployment.apps/")
}

func (kc KubeCommand) GetPods(selectors map[string]string) ([]string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name",
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "pod/")
}

func (kc KubeCommand) GetContainers(deployment v1.Deployment) []string {
	var containers []string
	for _, c := range deployment.Spec.Template.Spec.Containers {
		containers = append(containers, c.Name)
	}
	return containers
}

func (kc KubeCommand) GetPodsByLabels(labels []string) ([]string, error) {
	cmd := []string{
		kc.name, "get", "pods",
		"-l", strings.Join(labels, ","),
		"-o", "name", "-A",
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "pod/")
}

func (kc KubeCommand) RestartDeployment(deployment string) *Cmd {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"rollout", "restart", fmt.Sprintf("deployment/%v", deployment),
	}
	return Command(kc.l, cmd...)
}

func parseResources(out, delimiter, prefix string) ([]string, error) {
	var res []string
	if out == "" {
		return res, nil
	}
	lines := strings.Split(out, delimiter)
	if len(lines) == 1 && lines[0] == "" {
		return nil, fmt.Errorf("delimiter %q not found in %q", delimiter, out)
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

func (kc KubeCommand) CreateConfigMapFromFile(name, path string) (string, error) {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"create", "configmap", name,
		"--from-file", path,
	}
	return Command(kc.l, cmd...).Run()
}

func (kc KubeCommand) CreateConfigMap(name string, keyMap map[string]string) (string, error) {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"create", "configmap", name,
	}
	for key, value := range keyMap {
		cmd = append(cmd, fmt.Sprintf("--from-literal=%v=%v", key, value))
	}
	return Command(kc.l, cmd...).Run()
}

func (kc KubeCommand) DeleteConfigMap(name string) (string, error) {
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"delete", "configmap", name,
	}
	return Command(kc.l, cmd...).Run()
}

func (kc KubeCommand) GetConfigMapKey(name, key string) (string, error) {
	key = strings.ReplaceAll(key, ".", "\\.")
	// jsonpath map key is not very fond of dots
	cmd := []string{
		kc.name, "-n", kc.namespace,
		"get", "configmap", name,
		"-o", fmt.Sprintf("jsonpath={.data.%v}", key),
	}
	out, err := Command(kc.l, cmd...).Run()
	if err != nil {
		return out, err
	}
	if out == "" {
		return out, fmt.Errorf("no key %q found in ConfigMap %q", key, name)
	}
	return out, nil
}
