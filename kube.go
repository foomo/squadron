package configurd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

func rollbackDeployment(l *logrus.Entry, deployment, namespace string) *Cmd {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "undo", fmt.Sprintf("deployment/%v", deployment),
	}
	return command(l, cmd...)
}

func waitForRollout(l *logrus.Entry, deployment, namespace, timeout string) *Cmd {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "status", fmt.Sprintf("deployment/%v", deployment),
		"-w", "--timeout", timeout,
	}
	return command(l, cmd...)
}

func GetMostRecentPodBySelectors(l *logrus.Entry, selectors map[string]string, namespace string) (string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := []string{
		"kubectl", "-n", namespace,
		"--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name",
	}
	out, err := command(l, cmd...).run()
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

func waitForPodState(l *logrus.Entry, namepsace, pod, condition, timeout string) *Cmd {
	cmd := []string{
		"kubectl", "-n", namepsace,
		"wait", fmt.Sprintf("pod/%v", pod),
		fmt.Sprintf("--for=%v", condition),
		fmt.Sprintf("--timeout=%v", timeout),
	}
	return command(l, cmd...)
}

func execShell(l *logrus.Entry, resource, path, namespace string) *Cmd {
	cmdArgs := []string{
		"kubectl", "-n", namespace,
		"exec", "-it", resource,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd %v && /bin/sh", path),
	}

	return command(l, cmdArgs...).stdin(os.Stdin).stdout(os.Stdout).stderr(os.Stdout)
}

func patchDeployment(l *logrus.Entry, patch, deployment, namespace string) *Cmd {
	cmd := []string{
		"kubectl", "-n", namespace,
		"patch", "deployment", deployment,
		"--patch", patch,
	}
	return command(l, cmd...)
}

func copyToPod(l *logrus.Entry, pod, container, namespace, source, destination string) *Cmd {
	cmd := []string{
		"kubectl", "-n", namespace,
		"cp", source, fmt.Sprintf("%v:%v", pod, destination),
		"-c", container,
	}
	return command(l, cmd...)
}

func execPod(l *logrus.Entry, pod, container, namespace string, cmd []string) *Cmd {
	c := []string{
		"kubectl", "-n", namespace,
		"exec", pod,
		"-c", container,
		"--",
	}
	c = append(c, cmd...)
	return command(l, c...)
}

func exposePod(l *logrus.Entry, namespace, pod string, host string, port int) *Cmd {
	if host == "127.0.0.1" {
		host = ""
	}
	cmd := []string{
		"kubectl", "-n", namespace,
		"expose", "pod", pod,
		"--type=LoadBalancer",
		fmt.Sprintf("--port=%v", port),
		fmt.Sprintf("--external-ip=%v", host),
		// fmt.Sprintf("--name=%v-%v", pod, port),
	}
	return command(l, cmd...)
}

func deleteService(l *logrus.Entry, deployment *v1.Deployment, service string) *Cmd {
	cmd := []string{
		"kubectl", "-n", deployment.Namespace,
		"delete", "service", service,
	}
	return command(l, cmd...)
}

func GetDeployment(l *logrus.Entry, namespace, deployment string) (*v1.Deployment, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment", deployment,
		"-o", "json",
	}
	out, err := command(l, cmd...).run()
	if err != nil {
		return nil, err
	}
	var d v1.Deployment
	if err := json.Unmarshal([]byte(out), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func getNamespaces(l *logrus.Entry) ([]string, error) {
	cmd := []string{
		"kubectl",
		"get", "namespace",
		"-o", "name",
	}
	out, err := command(l, cmd...).run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "namespace/")
}

func getDeployments(l *logrus.Entry, namespace string) ([]string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment",
		"-o", "name",
	}
	out, err := command(l, cmd...).run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "deployment.apps/")
}

func getPods(l *logrus.Entry, namespace string, selectors map[string]string) ([]string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := []string{
		"kubectl", "-n", namespace,
		"--selector", strings.Join(selector, ","),
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name",
	}
	out, err := command(l, cmd...).run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "pod/")
}

func getContainers(l *logrus.Entry, deployment *v1.Deployment) []string {
	var containers []string
	for _, c := range deployment.Spec.Template.Spec.Containers {
		containers = append(containers, c.Name)
	}
	return containers
}

func getPodsByLabels(l *logrus.Entry, labels []string) ([]string, error) {
	var selector []string
	for k, v := range labels {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := []string{
		"kubectl", "get", "pods",
		"-l", strings.Join(labels, ","),
		"-o", "name", "-A",
	}
	out, err := command(l, cmd...).run()
	if err != nil {
		return nil, err
	}

	return parseResources(out, "\n", "pod/")
}

func parseResources(out, delimiter, prefix string) ([]string, error) {
	lines := strings.Split(out, delimiter)
	if len(lines) == 1 && lines[0] == "" {
		return nil, fmt.Errorf("delimiter %q not found in %q", delimiter, out)
	}
	var res []string
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
