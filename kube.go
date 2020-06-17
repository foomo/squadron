package configurd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strings"
	"time"

	"github.com/foomo/configurd/bindata"
	"github.com/go-delve/delve/service/rpc2"
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

func DelveCleanup(l *logrus.Entry, deployment *v1.Deployment, pod, container string) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping debug")
	}

	l.Infof("removing delve service")
	deleteService(l, deployment, pod)

	l.Infof("cleaning up debug processes")
	execPod(l, pod, container, deployment.Namespace, []string{"pkill", "-9", "dlv"})
	execPod(l, pod, container, deployment.Namespace, []string{"pkill", "-9", deployment.Name})
	return "", nil
}

func Delve(l *logrus.Entry, deployment *v1.Deployment, pod, container, input string, args []string, delveContinue bool, delvePort int) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping debug")
	}

	binPath := path.Join(os.TempDir(), deployment.Name)
	l.Infof("building %q for debug", input)
	_, err := debugBuild(l, input, binPath, []string{"GOOS=linux"})
	if err != nil {
		return "", err
	}

	l.Infof("copying binary to pod %v", pod)
	binDestination := fmt.Sprintf("/%v", deployment.Name)
	_, err = copyToPod(l, pod, container, deployment.Namespace, binPath, binDestination)
	if err != nil {
		return "", err
	}

	defer DelveCleanup(l, deployment, pod, container)
	signalCapture(l)

	// l.Infof("exposing deployment %v for delve", deployment.Name)
	// out, err := exposePod(l, deployment.Namespace, pod, delvePort)
	// if err != nil {
	// 	return out, err
	// }

	l.Infof("executing delve command on pod %v", pod)
	cmd := []string{
		"dlv", "exec", binDestination,
		"--api-version=2",
		"--headless",
		fmt.Sprintf("--listen=:%v", delvePort),
		"--accept-multiclient",
	}
	if delveContinue {
		cmd = append(cmd, "--continue")
	}
	if len(args) == 0 {
		args, err = getArgsFromPod(l, deployment.Namespace, pod, container)
		if err != nil {
			return "", err
		}
	}
	if len(args) > 0 {
		cmd = append(append(cmd, "--"), args...)
	}

	go postDelveRun(l, fmt.Sprintf(":%v", delvePort), 5)
	execPod(l, pod, container, deployment.Namespace, cmd)
	return "", nil
}

func Patch(l *logrus.Entry, deployment *v1.Deployment, container, image, tag, hostPath string) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if isPatched {
		l.Warnf("deployment already patched, running rollback first")
		out, err := Rollback(l, deployment)
		if err != nil {
			return out, err
		}
	}

	l.Infof("extracting dummy files")
	if err := bindata.RestoreAssets(os.TempDir(), "dummy"); err != nil {
		return "", err
	}
	dummyPath := path.Join(os.TempDir(), "dummy")

	l.Infof("building dummy image with %v:%v", image, tag)
	_, err := buildDummy(l, image, tag, dummyPath)
	if err != nil {
		return "", err
	}

	l.Infof("rendering deployment patch template")
	patch, err := renderTemplate(
		path.Join(dummyPath, devDeploymentPatchFile),
		newPatchValues(deployment.Name, container, hostPath),
	)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for deployment to get ready")
	out, err := waitForRollout(l, deployment.Name, deployment.Namespace, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("patching deployment for development")
	out, err = patchDeployment(l, patch, deployment.Name, deployment.Namespace)
	if err != nil {
		return out, err
	}

	l.Infof("getting most recent pod with selector from deployment %v", deployment.Name)
	pod, err := GetMostRecentPodBySelectors(l, deployment.Spec.Selector.MatchLabels, deployment.Namespace)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for pod %v with %q", pod, conditionContainersReady)
	out, err = waitForPodState(l, deployment.Namespace, pod, conditionContainersReady, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("copying deployment %v args into pod %v", deployment.Name, pod)
	if err := copyArgsToPod(l, deployment, pod, container); err != nil {
		return "", err
	}

	return "", nil
}

func Rollback(l *logrus.Entry, deployment *v1.Deployment) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping rollback")
	}

	l.Infof("rolling back deployment %v", deployment.Name)
	out, err := rollbackDeployment(l, deployment.Name, deployment.Namespace)
	if err != nil {
		return out, err
	}

	return "", nil
}

func Shell(l *logrus.Entry, deployment *v1.Deployment, pod string) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping shell exec")
	}

	l.Infof("waiting for pod %v with %q", pod, conditionContainersReady)
	out, err := waitForPodState(l, deployment.Namespace, pod, conditionContainersReady, defaultWaitTimeout)
	if err != nil {
		return out, err
	}

	l.Infof("running interactive shell for patched deployment %v", deployment.Name)
	return "", execShell(l, fmt.Sprintf("pod/%v", pod), getMountPath(deployment.Name), deployment.Namespace)
}

func deploymentIsPatched(l *logrus.Entry, deployment *v1.Deployment) bool {
	_, ok := deployment.Spec.Template.ObjectMeta.Labels[defaultPatchedLabel]
	return ok
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

func buildDummy(l *logrus.Entry, image, tag, path string) (string, error) {
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

func execPod(l *logrus.Entry, pod, container, namespace string, cmd []string) (string, error) {
	c := []string{
		"kubectl", "-n", namespace,
		"exec", pod,
		"-c", container,
		"--",
	}
	c = append(c, cmd...)
	return command(c...).run(l)
}

func exposePod(l *logrus.Entry, namespace, pod string, port int) (string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"expose", "pod", pod,
		"--type=LoadBalancer",
		fmt.Sprintf("--port=%v", port),
		// fmt.Sprintf("--name=%v-%v", pod, port),
	}
	return command(cmd...).run(l)
}

func deleteService(l *logrus.Entry, deployment *v1.Deployment, service string) (string, error) {
	cmd := []string{
		"kubectl", "-n", deployment.Namespace,
		"delete", "service", service,
	}
	return command(cmd...).run(l)
}

func getArgsFromPod(l *logrus.Entry, namespace, pod, container string) ([]string, error) {
	out, err := execPod(l, pod, container, namespace, []string{"cat", "/args.yaml"})
	if err != nil {
		return nil, err
	}
	var args []string
	if err := yaml.Unmarshal([]byte(out), &args); err != nil {
		return nil, err
	}
	return args, nil
}

func copyArgsToPod(l *logrus.Entry, deployment *v1.Deployment, pod, container string) error {
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
	_, err := copyToPod(l, pod, container, deployment.Namespace, argsSource, argsDestination)
	if err != nil {
		return err
	}
	return nil
}

func GetDeploymentImageTag(deployment *v1.Deployment, container string) (string, string, error) {
	for _, c := range deployment.Spec.Template.Spec.Containers {
		if c.Name == container {
			pieces := strings.Split(c.Image, ":")
			if len(pieces) != 2 {
				break
			}
			return pieces[0], pieces[1], nil
		}
	}
	return "", "", fmt.Errorf("could not find image from deployment and image flag not specified")
}

func getDeployment(l *logrus.Entry, namespace, deployment string) (*v1.Deployment, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment", deployment,
		"-o", "json",
	}
	out, err := command(cmd...).run(l)
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
	out, err := command(cmd...).run(l)
	if err != nil {
		return nil, err
	}

	var namespaces []string
	for _, n := range strings.Split(out, "\n") {
		namespaces = append(namespaces, strings.TrimPrefix(n, "namespace/"))
	}
	return namespaces, nil
}

func getDeployments(l *logrus.Entry, namespace string) ([]string, error) {
	cmd := []string{
		"kubectl", "-n", namespace,
		"get", "deployment",
		"-o", "name",
	}
	out, err := command(cmd...).run(l)
	if err != nil {
		return nil, err
	}

	var deployments []string
	for _, d := range strings.Split(out, "\n") {
		deployments = append(deployments, strings.TrimPrefix(d, "deployment.apps/"))
	}
	return deployments, nil
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
	out, err := command(cmd...).run(l)
	if err != nil {
		return nil, err
	}

	var pods []string
	for _, p := range strings.Split(out, "\n") {
		pods = append(pods, strings.TrimPrefix(p, "pod/"))
	}
	return pods, nil
}

func getContainers(l *logrus.Entry, deployment *v1.Deployment) []string {
	var containers []string
	for _, c := range deployment.Spec.Template.Spec.Containers {
		containers = append(containers, c.Name)
	}
	return containers
}

func validateResource(resourceType, resource string, available []string) error {
	if !stringInSlice(resource, available) {
		return fmt.Errorf("%v %v not found, available: %v", resourceType, resource, strings.Join(available, ", "))
	}
	return nil
}

func ValidateResources(l *logrus.Entry, namespace, deployment, pod, container string) (*v1.Deployment, error) {
	if err := validateNamespace(l, namespace); err != nil {
		return nil, err
	}
	if err := validateDeployment(l, namespace, deployment); err != nil {
		return nil, err
	}
	d, err := getDeployment(l, namespace, deployment)
	if err != nil {
		return nil, err
	}
	if err := validatePod(l, d, pod); pod != "" && err != nil {
		return nil, err
	}
	if err := validateContainer(l, d, container); err != nil {
		return nil, err
	}
	return d, nil
}

func validateNamespace(l *logrus.Entry, namespace string) error {
	available, err := getNamespaces(l)
	if err != nil {
		return err
	}
	return validateResource("namespace", namespace, available)
}

func validateDeployment(l *logrus.Entry, namespace, deployment string) error {
	available, err := getDeployments(l, namespace)
	if err != nil {
		return err
	}
	return validateResource("deployment", deployment, available)
}

func validatePod(l *logrus.Entry, deployment *v1.Deployment, pod string) error {
	available, err := getPods(l, deployment.Namespace, deployment.Spec.Selector.MatchLabels)
	if err != nil {
		return err
	}
	return validateResource("pod", pod, available)
}

func validateContainer(l *logrus.Entry, deployment *v1.Deployment, container string) error {
	available := getContainers(l, deployment)
	return validateResource("container", container, available)
}

func GetFreePort(host string) (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:0", host))
	if err != nil {
		return 0, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func signalCapture(l *logrus.Entry) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		l.Warnf("signal %s recieved", <-sigchan)
	}()
}

func checkDelveServer(l *logrus.Entry, addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		// l.WithError(err).Warnf("cannot connect to delve server on %v", addr)
		return err
	}
	client := rpc2.NewClientFromConn(conn)
	_, err = client.GetState()
	if err != nil {
		// l.WithError(err).Warnf("cannot get state from delve server on %v", addr)
		return err
	}
	return nil
}

func runOpen(l *logrus.Entry, path string) (string, error) {
	var cmd []string
	switch runtime.GOOS {
	case "linux":
		cmd = []string{"xdg-open", path}
	case "windows":
		cmd = []string{"rundll32", "url.dll,FileProtocolHandler", path}
	case "darwin":
		cmd = []string{"open", path}
	default:
		return "", fmt.Errorf("unsupported platform")
	}
	return command(cmd...).run(l)
}

func postDelveRun(l *logrus.Entry, addr string, maxTries int) {
	var err error
	for i := 0; i < maxTries; i++ {
		l.Infof("checking delve server on %v", addr)
		err = checkDelveServer(l, addr)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	if err != nil {
		l.WithError(err).Warn("failed connecting to delve server")
		return
	}

	l.Infof("opening debug configuration")
	out, err := runOpen(l, `vscode://fabiospampinato.vscode-debug-launcher/launch?args={"type":"node","name":"Foo","request":"launch","program":"/path/to/foo.js"}`)
	if err != nil {
		l.WithError(err).Warn(out)
	}
}
