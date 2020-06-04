package configurd

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"os/exec"
	"path"

	"github.com/foomo/configurd/exampledata"
	"github.com/sirupsen/logrus"
)

func RolloutDev(l *logrus.Entry, deployment, namespace, image, tag, mountPath string) (string, error) {
	l.Infof("extracting deployment patch file")
	if err := exampledata.RestoreAsset(os.TempDir(), devDeploymentPatchFile); err != nil {
		return "", err
	}
	patchPath := path.Join(os.TempDir(), devDeploymentPatchFile)

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
	out, err := waitForRollout(l, deployment, namespace, defaultRolloutTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("patching deployment for development")
	out, err = patchDeployment(l, patch, deployment, namespace)
	if err != nil {
		return out, err
	}
	defer rollbackDev(l, deployment, namespace, patchPath)

	l.Infof("waiting for deployment to get ready")
	out, err = waitForRollout(l, deployment, namespace, defaultRolloutTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("running interactive shell for patched deployment %v", deployment)
	return execDeploymentShell(l, deployment, namespace)
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

func execDeploymentShell(l *logrus.Entry, deployment, namespace string) (string, error) {
	cmdArgs := []string{
		"kubectl", "-n", namespace,
		"exec", "-it", fmt.Sprintf("deployment/%v", deployment),
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd /%v && /bin/sh", deployment),
	}
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr

	l.Tracef("executing %q from wd %q", cmd.String(), cmd.Dir)
	err := cmd.Run()
	if err != nil {
		// shouldnt return error since it triggers on exit from interactive shell
		// if the previous command has failed
		l.Warn(err)
	}
	return "", nil
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

func rollbackDev(l *logrus.Entry, deployment, namespace, patchPath string) {
	// called with defer
	l.Infof("removing deployment patch file")
	if err := os.Remove(patchPath); err != nil {
		l.WithError(err).Errorf("couldnt delete %v", patchPath)
	}

	l.Infof("rolling back deployment %v in namespace %v", deployment, namespace)
	_, err := rollbackDeployment(l, deployment, namespace)
	if err != nil {
		l.WithError(err).Errorf("deployment rollback failed")
	}
}
