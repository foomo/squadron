package configurd

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/foomo/configurd/exampledata"
	"github.com/sirupsen/logrus"
)

func (c Configurd) RolloutDev(service, namespace, image, tag, mountPath string) (string, error) {
	l := c.config.Log

	l.Infof("extracting deployment patch and template files")
	if err := exampledata.RestoreAsset(os.TempDir(), "deployment-patch.yaml"); err != nil {
		return "", err
	}
	if err := exampledata.RestoreAsset(os.TempDir(), "deployment-spec-selector.tpl"); err != nil {
		return "", err
	}
	patchPath := path.Join(os.TempDir(), "deployment-patch.yaml")
	templatePath := path.Join(os.TempDir(), "deployment-spec-selector.tpl")

	l.Infof("rendering deployment patch template")
	patch, err := renderTemplate(
		patchPath,
		map[string]string{
			"Name":      service,
			"MountPath": fmt.Sprintf("/%v", service),
			"HostPath":  mountPath,
			"Image":     fmt.Sprintf("%v:%v", image, tag),
		},
	)
	if err != nil {
		return "", err
	}

	l.Infof("patching deployment for development")
	out, err := patchDeployment(l, patch, service, namespace)
	if err != nil {
		return out, err
	}
	defer rollbackDev(l, service, namespace, patchPath, templatePath)

	l.Infof("getting selectors for deployment %v in namespace %v", service, namespace)
	selectors, err := getDeploymentSelectors(l, service, namespace, templatePath)
	if err != nil {
		return out, err
	}

	l.Infof("getting most recent pod names from deployment %v in namespace %v", service, namespace)
	out, err = getMostRecentPodName(l, selectors, namespace)
	if err != nil {
		return out, err
	}
	podName := out

	l.Infof("waiting for pod %v to get ready", podName)
	out, err = waitForPodState(l, namespace, podName, "condition=Ready", "30s")
	if err != nil {
		return out, err
	}

	l.Infof("running interactive shell for patched pod %v", podName)
	out, err = execPodShell(l, podName, service, namespace)
	spew.Dump(out, err)

	return "", nil
}

func rollbackDeployment(l *logrus.Entry, service, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "undo", fmt.Sprintf("deployment/%v", service),
	}
	return runCommand(l, "", nil, cmd...)
}

func getDeploymentSelectors(l *logrus.Entry, name, namespace, templatePath string) (map[string]string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment", name,
		"-o", fmt.Sprintf("go-template-file=%v", templatePath),
	}
	out, err := runCommand(l, "", nil, cmd...)
	if err != nil {
		return nil, fmt.Errorf("%v, error:%s", out, err)
	}

	selectors := make(map[string]string)
	for _, s := range strings.Split(out, "\n") {
		pieces := strings.Split(s, ":")
		selectors[pieces[0]] = pieces[1]
	}
	return selectors, nil
}

func getMostRecentPodName(l *logrus.Entry, selectors map[string]string, namespace string) (string, error) {
	var selector []string
	for k, v := range selectors {
		selector = append(selector, fmt.Sprintf("%v=%v", k, v))
	}
	cmd := []string{
		"kubectl", "-n", namespace,
		"--selector", strings.Join(selector, ","),
		"--field-selector", "status.phase=Pending",
		"get", "pods", "--sort-by=.status.startTime",
		"-o", "name",
	}
	out, err := runCommand(l, "", nil, cmd...)
	if err != nil {
		return out, err
	}
	pods := strings.Split(out, "\n")
	pod := pods[0]
	if len(pods) > 1 {
		pod = pods[len(pods)-1]
	}
	if pod == "" {
		return "", fmt.Errorf("no pods found")
	}
	return pod, nil
}

func waitForPodState(l *logrus.Entry, namepsace, pod, condition, timeout string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namepsace, "wait",
		fmt.Sprintf("--for=%v", condition), pod,
		fmt.Sprintf("--timeout=%v", timeout),
	}
	return runCommand(l, "", nil, cmd...)
}

func execPodShell(l *logrus.Entry, pod, service, namespace string) (string, error) {
	cmdArgs := []string{
		"kubectl", "-n", namespace,
		"exec", "-it", pod,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd /%v && /bin/sh", service),
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()

	return "", err
}

func patchDeployment(l *logrus.Entry, patch, name, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "patch", "deployment", name,
		"-n", namespace,
		"--patch", patch,
	}
	return runCommand(l, "", nil, cmd...)
}

func renderTemplate(path string, values interface{}) (string, error) {
	tpl, err := template.ParseFiles(path)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tpl.Execute(buf, values)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func rollbackDev(l *logrus.Entry, service, namespace, patchPath, templatePath string) {
	l.Infof("removing deployment patch and template files")
	if err := os.Remove(patchPath); err != nil {
		l.WithError(err).Warnf("couldnt delete %v", patchPath)
	}
	if err := os.Remove(templatePath); err != nil {
		l.WithError(err).Warnf("couldnt delete %v", templatePath)
	}

	l.Infof("rolling back deployment %v in namespace %v", service, namespace)
	_, err := rollbackDeployment(l, service, namespace)
	if err != nil {
		l.WithError(err).Warnf("deployment rollback failed")
	}
}
