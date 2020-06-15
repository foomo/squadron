package configurd

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path"
	"strings"
	"time"

	"github.com/foomo/configurd/bindata"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v1"
	v1 "k8s.io/api/apps/v1"
)

type patchValues struct {
	PatchedLabelName string
	ContainerName    string
	MountPath        string
	HostPath         string
	Image            string
}

func newPatchValues(deployment, container, hostPath string) *patchValues {
	return &patchValues{
		PatchedLabelName: defaultPatchedLabel,
		ContainerName:    container,
		MountPath:        getMountPath(deployment),
		HostPath:         hostPath,
		Image:            "dummy:latest",
	}
}

func DelveDevCleanup(l *logrus.Entry, namespace, deployment, container string) (string, error) {
	l.Infof("checking if deployment is patched")
	isPatched, err := deploymentIsPatched(l, deployment, namespace)
	if err != nil {
		return "", err
	}
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping debug")
	}

	l.Infof("getting deployment %v info", deployment)
	d, err := getDeployment(l, deployment, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("getting most recent pod with selector from deployment %v", deployment)
	pod, err := getMostRecentPodBySelectors(l, d.Spec.Selector.MatchLabels, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("cleaning up debug processes")
	execPod(l, pod, container, namespace, []string{"pkill", "dlv"}, 0)
	execPod(l, pod, container, namespace, []string{"pkill", deployment}, 0)
	return "", nil
}

func DelveDev(l *logrus.Entry, namespace, deployment, container, input string, args []string) (string, error) {
	l.Infof("checking if deployment is patched")
	isPatched, err := deploymentIsPatched(l, deployment, namespace)
	if err != nil {
		return "", err
	}
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping debug")
	}

	binPath := path.Join(os.TempDir(), deployment)
	l.Infof("building %q for debug", input)
	_, err = debugBuild(l, input, binPath, []string{"GOOS=linux"})
	if err != nil {
		return "", err
	}

	l.Infof("getting deployment %v info", deployment)
	d, err := getDeployment(l, deployment, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("getting most recent pod with selector from deployment %v", deployment)
	pod, err := getMostRecentPodBySelectors(l, d.Spec.Selector.MatchLabels, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("copying binary to pod %v", pod)
	binDestination := fmt.Sprintf("/%v", deployment)
	_, err = copyToPod(l, pod, container, namespace, binPath, binDestination)
	if err != nil {
		return "", err
	}

	l.Infof("executing delve command on pod %v", pod)
	cmd := []string{
		"dlv", "exec", binDestination,
		"--api-version=2",
		"--headless",
		"--listen=:2345",
		// "--accept-multiclient",
		// "--continue",
	}
	if len(args) == 0 {
		args, err = getArgsFromPod(l, namespace, pod, container)
		if err != nil {
			return "", err
		}
	}
	if len(args) > 0 {
		cmd = append(append(cmd, "--"), args...)
	}
	return execPod(l, pod, container, namespace, cmd, 1)
}

func RolloutDev(l *logrus.Entry, deployment, container, namespace, image, tag, hostPath string, goDebug bool) (string, error) {
	l.Infof("checking if deployment is patched")
	isPatched, err := deploymentIsPatched(l, deployment, namespace)
	if err != nil {
		return "", err
	}
	if isPatched {
		return "", fmt.Errorf("deployment already patched, to patch again, run stop first")
	}

	l.Infof("extracting dummy files")
	if err := bindata.RestoreAssets(os.TempDir(), "dummy"); err != nil {
		return "", err
	}
	dummyPath := path.Join(os.TempDir(), "dummy")

	l.Infof("building dummy image with %v:%v", image, tag)
	_, err = buildDummy(l, image, tag, dummyPath, goDebug)
	if err != nil {
		return "", err
	}

	l.Infof("getting container names for deployment %v", deployment)
	containers, err := getContainerNames(l, deployment, namespace)
	if err != nil {
		return "", err
	}
	if !stringInSlice(container, containers) {
		return "", fmt.Errorf("Could not find container %v defined in deployment %v, available: %v",
			container, deployment, strings.Join(containers, ", "))
	}

	l.Infof("rendering deployment patch template")
	patch, err := renderTemplate(
		path.Join(dummyPath, devDeploymentPatchFile),
		newPatchValues(deployment, container, hostPath),
	)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for deployment to get ready")
	out, err := waitForRollout(l, deployment, namespace, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("getting deployment %v info", deployment)
	d, err := getDeployment(l, deployment, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("patching deployment for development")
	out, err = patchDeployment(l, patch, deployment, namespace)
	if err != nil {
		return out, err
	}

	if goDebug {
		l.Infof("exposing deployment %v for delve", deployment)
		out, err = exposeDeployment(l, deployment, namespace, 2345)
		if err != nil {
			return out, err
		}
	}

	l.Infof("getting most recent pod with selector from deployment %v", deployment)
	pod, err := getMostRecentPodBySelectors(l, d.Spec.Selector.MatchLabels, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for pod %v with %q", pod, conditionContainersReady)
	out, err = waitForPodState(l, namespace, pod, conditionContainersReady, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("copying deployment %v args into pod %v", deployment, pod)
	if err := copyArgsToPod(l, d, namespace, pod, container); err != nil {
		return "", err
	}

	return "", nil
}

func RollbackDev(l *logrus.Entry, deployment, namespace string) (string, error) {
	l.Infof("checking if deployment is patched")
	isPatched, err := deploymentIsPatched(l, deployment, namespace)
	if err != nil {
		return "", err
	}
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping rollback")
	}

	l.Infof("rolling back deployment %v", deployment)
	out, err := rollbackDeployment(l, deployment, namespace)
	if err != nil {
		return out, err
	}

	l.Infof("removing delve service")
	out, err = deleteDelveService(l, deployment, namespace)
	if err != nil {
		//may not exist
		l.WithError(err).Warnf(out)
	}

	return "", nil
}

func ShellDev(l *logrus.Entry, deployment, namespace string) (string, error) {
	l.Infof("checking if deployment is patched")
	isPatched, err := deploymentIsPatched(l, deployment, namespace)
	if err != nil {
		return "", err
	}
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping shell exec")
	}

	l.Infof("getting deployment %v info", deployment)
	d, err := getDeployment(l, deployment, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("getting most recent pod with selector from deployment %v", deployment)
	pod, err := getMostRecentPodBySelectors(l, d.Spec.Selector.MatchLabels, namespace)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for pod %v with %q", pod, conditionContainersReady)
	out, err := waitForPodState(l, namespace, pod, conditionContainersReady, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("running interactive shell for patched deployment %v", deployment)
	return "", execShell(l, fmt.Sprintf("pod/%v", pod), getMountPath(deployment), namespace)
}

func deploymentIsPatched(l *logrus.Entry, deployment, namespace string) (bool, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment", deployment,
		"-o", fmt.Sprintf("jsonpath={.spec.template.metadata.labels.%v}", defaultPatchedLabel),
	}
	out, err := command(cmd...).run(l)
	if err != nil {
		return false, err
	}
	if out == "true" {
		return true, nil
	}
	return false, nil
}

func rollbackDeployment(l *logrus.Entry, deployment, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "undo", fmt.Sprintf("deployment/%v", deployment),
	}
	return command(cmd...).run(l)
}

func waitForRollout(l *logrus.Entry, deployment, namespace, timeout string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"rollout", "status", fmt.Sprintf("deployment/%v", deployment),
		"-w", "--timeout", timeout,
	}
	return command(cmd...).run(l)
}

func getContainerNames(l *logrus.Entry, deployment, namespace string) ([]string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment", deployment,
		"-o", fmt.Sprintf("jsonpath=%v", "{.spec.template.spec.containers[*].name}"),
	}
	out, err := command(cmd...).run(l)
	if err != nil {
		return nil, fmt.Errorf("%v, error:%s", out, err)
	}

	return strings.Split(out, " "), nil
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
	out, err := command(cmd...).run(l)
	if err != nil {
		return out, err
	}
	pods := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
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
	return command(cmd...).run(l)
}

func execShell(l *logrus.Entry, resource, path, namespace string) error {
	cmdArgs := []string{
		"kubectl", "-n", namespace,
		"exec", "-it", resource,
		"--", "/bin/sh", "-c",
		fmt.Sprintf("cd %v && /bin/sh", path),
	}

	_, err := command(cmdArgs...).stdin(os.Stdin).stdout(os.Stdout).stderr(os.Stdout).run(l)
	if err != nil {
		l.Warn(err)
	}
	return nil
}

func patchDeployment(l *logrus.Entry, patch, deployment, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"patch", "deployment", deployment,
		"--patch", patch,
	}
	return command(cmd...).run(l)
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

func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func getMountPath(name string) string {
	return fmt.Sprintf("/%v-mount", name)
}

func buildDummy(l *logrus.Entry, image, tag, path string, goDebug bool) (string, error) {
	if goDebug {
		l.Warnf("go debug mode override: using golang:alpine as image")
		image = "golang"
		tag = "alpine"
	}
	cmd := []string{
		"docker", "build", ".",
		"--build-arg", fmt.Sprintf("IMAGE=%v:%v", image, tag),
		"-t", "dummy:latest",
	}
	return command(cmd...).cwd(path).run(l)
}

func debugBuild(l *logrus.Entry, input, output string, env []string) (string, error) {
	cmd := []string{
		"go", "build",
		"-gcflags=\"all=-N -l\"",
		"-o", output, input,
	}
	return command(cmd...).env(env).run(l)
}

func copyToPod(l *logrus.Entry, pod, container, namespace, source, destination string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"cp", source, fmt.Sprintf("%v:%v", pod, destination),
		"-c", container,
	}
	return command(cmd...).run(l)
}

func execPod(l *logrus.Entry, pod, container, namespace string, cmd []string, timeout time.Duration) (string, error) {
	c := []string{
		"kubectl", "-n", namespace,
		"exec", pod,
		"-c", container,
		"--",
	}
	c = append(c, cmd...)
	return command(c...).timeout(timeout).run(l)
}

func exposeDeployment(l *logrus.Entry, deployment, namespace string, port int) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"expose", "deployment", deployment,
		"--type=LoadBalancer",
		fmt.Sprintf("--port=%v", port),
		fmt.Sprintf("--name=%v-delve", deployment),
	}
	return command(cmd...).run(l)
}

func deleteDelveService(l *logrus.Entry, deployment, namespace string) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"delete", "service",
		fmt.Sprintf("%v-delve", deployment),
	}
	return command(cmd...).run(l)
}

func getDeploymentArgs(l *logrus.Entry, deployment *v1.Deployment, container string) ([]string, error) {
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			return c.Args, nil
		}
	}
	return nil, fmt.Errorf("deployment %v args not found", deployment)
}

func getDeployment(l *logrus.Entry, deployment, namespace string) (*v1.Deployment, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment", deployment,
		"-o", "yaml",
	}
	out, err := command(cmd...).run(l)
	if err != nil {
		return nil, err
	}
	var d v1.Deployment
	if err := yaml.Unmarshal([]byte(out), &d); err != nil {
		return nil, err
	}
	return &d, nil
}

func getArgsFromPod(l *logrus.Entry, namespace, pod, container string) ([]string, error) {
	out, err := execPod(l, pod, container, namespace, []string{"cat", "/args.yaml"}, 0)
	if err != nil {
		return nil, err
	}
	var args []string
	if err := yaml.Unmarshal([]byte(out), &args); err != nil {
		return nil, err
	}
	return args, nil
}

func copyArgsToPod(l *logrus.Entry, deployment *v1.Deployment, namespace, pod, container string) error {
	var args []string
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			args = c.Args
			break
		}
	}

	argsSource := path.Join(os.TempDir(), "args.yaml")
	if err := generateYaml(l, argsSource, args); err != nil {
		return err
	}
	argsDestination := "/args.yaml"
	_, err := copyToPod(l, pod, container, namespace, argsSource, argsDestination)
	if err != nil {
		return err
	}
	return nil
}
