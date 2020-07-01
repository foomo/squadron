package util

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

func NewKubeCommand(l *logrus.Entry) (*CliCommand, error) {
	return NewCliCommand(l, "kubectl")
}

func (cc CliCommand) RollbackDeployment(deployment string) *Cmd {
	cmd := append(cc.GetCommand(), "rollout", "undo", fmt.Sprintf("deployment/%v", deployment))
	return Command(cc.l, cmd...)
}

func (cc CliCommand) WaitForRollout(deployment, timeout string) *Cmd {
	cmd := append(cc.GetCommand(),
		"rollout", "status",
		fmt.Sprintf("deployment/%v", deployment),
		"-w", "--timeout", timeout)
	return Command(cc.l, cmd...)
}

func (cc CliCommand) GetMostRecentPodBySelectors(selectors map[string]string) (string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := append(cc.GetCommand(),
		"--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name")
	out, err := Command(cc.l, cmd...).Run()
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

func (cc CliCommand) WaitForPodState(pod, condition, timeout string) *Cmd {
	cmd := append(cc.GetCommand(),
		"wait", fmt.Sprintf("pod/%v", pod),
		fmt.Sprintf("--for=%v", condition),
		fmt.Sprintf("--timeout=%v", timeout))
	return Command(cc.l, cmd...)
}

func (cc CliCommand) ExecShell(resource, path string) *Cmd {
	cmd := append(cc.GetCommand(),
		"exec", "-it", resource,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd %v && /bin/sh", path))
	return Command(cc.l, cmd...).Stdin(os.Stdin).Stdout(os.Stdout).Stderr(os.Stdout)
}

func (cc CliCommand) PatchDeployment(patch, deployment string) *Cmd {
	cmd := append(cc.GetCommand(),
		"patch", "deployment", deployment,
		"--patch", patch)
	return Command(cc.l, cmd...)
}

func (cc CliCommand) CopyToPod(pod, container, source, destination string) *Cmd {
	cmd := append(cc.GetCommand(),
		"cp", source, fmt.Sprintf("%v:%v", pod, destination),
		"-c", container)
	return Command(cc.l, cmd...)
}

func (cc CliCommand) ExecPod(pod, container string, cmd []string) *Cmd {
	c := append(cc.GetCommand(),
		"exec", pod,
		"-c", container, "--")
	c = append(c, cmd...)
	return Command(cc.l, c...)
}

func (cc CliCommand) ExposePod(pod string, host string, port int) *Cmd {
	if host == "127.0.0.1" {
		host = ""
	}
	cmd := append(cc.GetCommand(),
		"expose", "pod", pod,
		"--type=LoadBalancer",
		fmt.Sprintf("--port=%v", port),
		fmt.Sprintf("--external-ip=%v", host))
	return Command(cc.l, cmd...)
}

func (cc CliCommand) DeleteService(service string) *Cmd {
	cmd := append(cc.GetCommand(),
		"delete", "service", service)
	return Command(cc.l, cmd...)
}

func (cc CliCommand) GetDeployment(deployment string) (*v1.Deployment, error) {
	cmd := append(cc.GetCommand(),
		"get", "deployment", deployment,
		"-o", "json")
	out, err := Command(cc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}
	var d v1.Deployment
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func (cc CliCommand) GetNamespaces() ([]string, error) {
	cmd := append(cc.GetCommand(),
		"get", "namespace",
		"-o", "name")
	out, err := Command(cc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}
	return parseResources(out, "\n", "namespace/")
}

func (cc CliCommand) GetDeployments() ([]string, error) {
	cmd := append(cc.GetCommand(),
		"get", "deployment",
		"-o", "name")
	out, err := Command(cc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}
	return parseResources(out, "\n", "deployment.apps/")
}

func (cc CliCommand) GetPods(selectors map[string]string) ([]string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := append(cc.GetCommand(),
		"--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name")
	out, err := Command(cc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}
	return parseResources(out, "\n", "pod/")
}

func (cc CliCommand) GetContainers(deployment v1.Deployment) []string {
	var containers []string
	for _, c := range deployment.Spec.Template.Spec.Containers {
		containers = append(containers, c.Name)
	}
	return containers
}

func (cc CliCommand) GetPodsByLabels(labels []string) ([]string, error) {
	cmd := append(cc.GetCommand(),
		"get", "pods",
		"-l", strings.Join(labels, ","),
		"-o", "name", "-A")
	out, err := Command(cc.l, cmd...).Run()
	if err != nil {
		return nil, err
	}
	return parseResources(out, "\n", "pod/")
}

func (cc CliCommand) RestartDeployment(deployment string) *Cmd {
	cmd := append(cc.GetCommand(),
		"rollout", "restart", fmt.Sprintf("deployment/%v", deployment))
	return Command(cc.l, cmd...)
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

func (cc CliCommand) CreateConfigMapFromFile(name, path string) (string, error) {
	cmd := append(cc.GetCommand(),
		"create", "configmap", name,
		"--from-file", path)
	return Command(cc.l, cmd...).Run()
}

func (cc CliCommand) CreateConfigMap(name string, keyMap map[string]string) (string, error) {
	cmd := append(cc.GetCommand(),
		"create", "configmap", name)
	for key, value := range keyMap {
		cmd = append(cmd, fmt.Sprintf("--from-literal=%v=%v", key, value))
	}
	return Command(cc.l, cmd...).Run()
}

func (cc CliCommand) DeleteConfigMap(name string) (string, error) {
	cmd := append(cc.GetCommand(),
		"delete", "configmap", name)
	return Command(cc.l, cmd...).Run()
}

func (cc CliCommand) GetConfigMapKey(name, key string) (string, error) {
	key = strings.ReplaceAll(key, ".", "\\.")
	// jsonpath map key is not very fond of dots
	cmd := append(cc.GetCommand(),
		"get", "configmap", name,
		"-o", fmt.Sprintf("jsonpath={.data.%v}", key))
	out, err := Command(cc.l, cmd...).Run()
	if err != nil {
		return out, err
	}
	if out == "" {
		return out, fmt.Errorf("no key %q found in ConfigMap %q", key, name)
	}
	return out, nil
}
