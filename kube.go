package configurd

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/foomo/configurd/exampledata"
	"github.com/sirupsen/logrus"
)

func RolloutDev(l *logrus.Entry, deployment, namespace, image, tag, mountPath string) (string, error) {
	l.Infof("extracting deployment patch file")
	if err := exampledata.RestoreAsset(os.TempDir(), devDeploymentPatchFile); err != nil {
		return "", err
	}
	patchPath := path.Join(os.TempDir(), devDeploymentPatchFile)

	l.Infof("extracting deployment template file")
	if err := exampledata.RestoreAsset(os.TempDir(), devDeploymentTemplateFile); err != nil {
		return "", err
	}
	templatePath := path.Join(os.TempDir(), devDeploymentTemplateFile)

	l.Infof("rendering deployment patch template")
	patch, err := renderTemplate(
		patchPath,
		map[string]string{
			"Name":      deployment,
			"MountPath": fmt.Sprintf("/%v", deployment),
			"HostPath":  mountPath,
			"Image":     fmt.Sprintf("%v:%v", image, tag),
		},
	)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for deployment to get ready")
	out, err := waitForRollout(l, deployment, namespace, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("patching deployment for development")
	out, err = patchDeployment(l, patch, deployment, namespace)
	if err != nil {
		return out, err
	}
	defer rollbackDev(l, deployment, namespace, patchPath, templatePath)

	l.Infof("getting selectors for deployment %v in namespace %v", deployment, namespace)
	selectors, err := getDeploymentSelectors(l, deployment, namespace, templatePath)
	if err != nil {
		return out, err
	}

	l.Infof("getting most recent pod names from deployment %v in namespace %v", deployment, namespace)
	pod, err := getMostRecentPodBySelectors(l, selectors, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for pod %v with %q", pod, conditionContainersReady)
	out, err = waitForPodState(l, namespace, pod, conditionContainersReady, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("running interactive shell for patched deployment %v", deployment)
	return execShell(l, fmt.Sprintf("pod/%v", pod), fmt.Sprintf("/%v", deployment), namespace)
}

func rollbackDeployment(l *logrus.Entry, deployment, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "undo", fmt.Sprintf("deployment/%v", deployment),
	}
	return runCommand(l, "", nil, cmd...)
}

func waitForRollout(l *logrus.Entry, deployment, namespace, timeout string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "status", fmt.Sprintf("deployment/%v", deployment),
		"-w", "--timeout", timeout,
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
	for _, s := range strings.Split(strings.TrimSuffix(out, "\n"), "\n") {
		pieces := strings.Split(s, ":")
		selectors[pieces[0]] = pieces[1]
	}
	return selectors, nil
}

func getMostRecentPodBySelectors(l *logrus.Entry, selectors map[string]string, namespace string) (string, error) {
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
	return strings.TrimPrefix(pod, "pod/"), nil
}

func waitForPodState(l *logrus.Entry, namepsace, pod, condition, timeout string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namepsace,
		"wait", fmt.Sprintf("pod/%v", pod),
		fmt.Sprintf("--for=%v", condition),
		fmt.Sprintf("--timeout=%v", timeout),
	}
	return runCommand(l, "", nil, cmd...)
}

func execShell(l *logrus.Entry, resource, path, namespace string) (string, error) {
	cmdArgs := []string{
		"kubectl", "-n", namespace,
		"exec", "-it", resource,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd /%v && /bin/sh", path),
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	l.Tracef("executing %q from wd %q", cmd.String(), cmd.Dir)
	err := cmd.Run()
	if err != nil {
		// shouldnt return error since it triggers
		// on exit from interactive shell
		// if the previous command has failed
		l.Warn(err)
	}
	return "", nil
}

func patchDeployment(l *logrus.Entry, patch, deployment, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"patch", "deployment", deployment,
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

func rollbackDev(l *logrus.Entry, deployment, namespace, patchPath, templatePath string) {
	// called with defer
	l.Infof("removing deployment patch file")
	if err := os.Remove(patchPath); err != nil {
		l.WithError(err).Errorf("couldnt delete %v", patchPath)
	}

	l.Infof("removing deployment template file")
	if err := os.Remove(templatePath); err != nil {
		l.WithError(err).Errorf("couldnt delete %v", templatePath)
	}

	l.Infof("rolling back deployment %v in namespace %v", deployment, namespace)
	_, err := rollbackDeployment(l, deployment, namespace)
	if err != nil {
		l.WithError(err).Errorf("deployment rollback failed")
	}
}
