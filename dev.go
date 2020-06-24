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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/foomo/configurd/bindata"
	"github.com/go-delve/delve/service/rpc2"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v1"
	v1 "k8s.io/api/apps/v1"
)

type Mount struct {
	HostPath  string
	MountPath string
}

type patchValues struct {
	PatchedLabelName string
	ContainerName    string
	Mounts           []Mount
	Image            string
}

func newPatchValues(container string, mounts []Mount) *patchValues {
	return &patchValues{
		PatchedLabelName: defaultPatchedLabel,
		ContainerName:    container,
		Mounts:           mounts,
		Image:            "dummy:latest",
	}
}

type launchArgs struct {
	Name       string `json:"name,omitempty"`
	Request    string `json:"request,omitempty"`
	Type       string `json:"type,omitempty"`
	Mode       string `json:"mode,omitempty"`
	RemotePath string `json:"remotePath,omitempty"`
	Port       int    `json:"port,omitempty"`
	Host       string `json:"host,omitempty"`
	Trace      string `json:"trace,omitempty"`
	LogOutput  string `json:"logOutput,omitempty"`
	ShowLog    bool   `json:"showLog,omitempty"`
}

func newLaunchArgs(pod, host string, port int) *launchArgs {
	return &launchArgs{
		Host:       host,
		Name:       fmt.Sprintf("delve-%v", pod),
		Port:       port,
		Request:    "attach",
		Type:       "go",
		Mode:       "remote",
		RemotePath: "${workspaceFolder}",
		// Trace:      "verbose",
		// LogOutput: "rpc",
		// ShowLog:   true,
	}
}

func (la *launchArgs) toJson() (string, error) {
	bytes, err := json.Marshal(la)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func DelveCleanup(l *logrus.Entry, deployment *v1.Deployment, pod, container string) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping debug")
	}

	l.Infof("removing delve service")
	deleteService(l, deployment, pod).run()

	l.Infof("cleaning up debug processes")
	execPod(l, pod, container, deployment.Namespace, []string{"pkill", "-9", "dlv"}).run()
	execPod(l, pod, container, deployment.Namespace, []string{"pkill", "-9", deployment.Name}).run()
	return "", nil
}

func Delve(l *logrus.Entry, deployment *v1.Deployment, pod, container, input string, args []string, delveContinue bool, host string, port int, vscode bool) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if !isPatched {
		return "", fmt.Errorf("deployment not patched, stopping debug")
	}

	goModDir, err := findGoProjectRoot(input)
	if err != nil {
		return "", fmt.Errorf("couldnt find go.mod dir for input %q", input)
	}

	binPath := path.Join(os.TempDir(), deployment.Name)
	l.Infof("building %q for debug", input)
	_, err = debugBuild(l, input, goModDir, binPath, []string{"GOOS=linux"})
	if err != nil {
		return "", err
	}

	l.Infof("copying binary to pod %v", pod)
	binDestination := fmt.Sprintf("/%v", deployment.Name)
	_, err = copyToPod(l, pod, container, deployment.Namespace, binPath, binDestination).run()
	if err != nil {
		return "", err
	}

	defer DelveCleanup(l, deployment, pod, container)
	signalCapture(l)

	l.Infof("exposing deployment %v for delve", deployment.Name)
	out, err := exposePod(l, deployment.Namespace, pod, host, port).run()
	if err != nil {
		return out, err

	}

	l.Infof("executing delve command on pod %v", pod)
	cmd := []string{
		"dlv", "exec", binDestination,
		"--api-version=2", "--headless",
		fmt.Sprintf("--listen=:%v", port),
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
		cmd = append(cmd, "--")
		cmd = append(cmd, args...)
	}

	execPod(l, pod, container, deployment.Namespace, cmd).postStart(
		func() error {
			if err := tryDelveServer(l, host, port, 5, 1*time.Second); err != nil {
				return err
			}
			if vscode {
				if err := launchVscode(l, goModDir, pod, host, port, 5, 1*time.Second); err != nil {
					return err
				}
			}
			return nil
		},
	).run()
	return "", nil
}

func Patch(l *logrus.Entry, deployment *v1.Deployment, container, image, tag string, mounts []Mount) (string, error) {
	isPatched := deploymentIsPatched(l, deployment)
	if isPatched {
		l.Warnf("deployment already patched, running rollback first")
		out, err := Rollback(l, deployment)
		if err != nil {
			return out, err
		}
	}

	l.Infof("waiting for deployment to get ready")
	out, err := waitForRollout(l, deployment.Name, deployment.Namespace, defaultWaitTimeout).run()
	if err != nil {
		return out, err
	}

	l.Infof("extracting dummy files")
	if err := bindata.RestoreAssets(os.TempDir(), "dummy"); err != nil {
		return "", err
	}
	dummyPath := path.Join(os.TempDir(), "dummy")

	l.Infof("building dummy image with %v:%v", image, tag)
	_, err = buildDummy(l, image, tag, dummyPath)
	if err != nil {
		return "", err
	}

	l.Infof("rendering deployment patch template")
	patch, err := renderTemplate(
		path.Join(dummyPath, devDeploymentPatchFile),
		newPatchValues(container, mounts),
	)
	if err != nil {
		return "", err
	}

	l.Infof("patching deployment for development")
	out, err = patchDeployment(l, patch, deployment.Name, deployment.Namespace).run()
	if err != nil {
		return out, err
	}

	l.Infof("getting most recent pod with selector from deployment %v", deployment.Name)
	pod, err := GetMostRecentPodBySelectors(l, deployment.Spec.Selector.MatchLabels, deployment.Namespace)
	if err != nil {
		return "", err
	}

	l.Infof("waiting for pod %v with %q", pod, conditionContainersReady)
	out, err = waitForPodState(l, deployment.Namespace, pod, conditionContainersReady, defaultWaitTimeout).run()
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

	l.Infof("waiting for deployment to get ready")
	out, err := waitForRollout(l, deployment.Name, deployment.Namespace, defaultWaitTimeout).run()
	if err != nil {
		return out, err
	}

	l.Infof("rolling back deployment %v", deployment.Name)
	out, err = rollbackDeployment(l, deployment.Name, deployment.Namespace).run()
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
	out, err := waitForPodState(l, deployment.Namespace, pod, conditionContainersReady, defaultWaitTimeout).run()
	if err != nil {
		return out, err
	}

	l.Infof("running interactive shell for patched deployment %v", deployment.Name)
	return execShell(l, fmt.Sprintf("pod/%v", pod), "/", deployment.Namespace).run()
}

func FindFreePort(host string) (int, error) {
	tcpAddr, err := CheckTCPConnection(host, 0)
	if err != nil {
		return 0, err
	}
	return tcpAddr.Port, nil
}

func CheckTCPConnection(host string, port int) (*net.TCPAddr, error) {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%v:%v", host, port))
	if err != nil {
		return nil, err
	}
	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr), nil
}

func deploymentIsPatched(l *logrus.Entry, deployment *v1.Deployment) bool {
	_, ok := deployment.Spec.Template.ObjectMeta.Labels[defaultPatchedLabel]
	return ok
}

func validateResource(resourceType, resource, suffix string, available []string) error {
	if !stringInSlice(resource, available) {
		return fmt.Errorf("%v %q not found %v, available: %v", resourceType, resource, suffix, strings.Join(available, ", "))
	}
	return nil
}

func ValidateNamespace(l *logrus.Entry, namespace string) error {
	available, err := getNamespaces(l)
	if err != nil {
		return err
	}
	return validateResource("namespace", namespace, "", available)
}

func ValidateDeployment(l *logrus.Entry, namespace, deployment string) error {
	available, err := getDeployments(l, namespace)
	if err != nil {
		return err
	}
	return validateResource("deployment", deployment, fmt.Sprintf("for namespace %q", namespace), available)
}

func ValidatePod(l *logrus.Entry, deployment *v1.Deployment, pod *string) error {
	if *pod == "" {
		var err error
		*pod, err = GetMostRecentPodBySelectors(l, deployment.Spec.Selector.MatchLabels, deployment.Namespace)
		if err != nil || *pod == "" {
			return err
		}
		return nil
	}
	available, err := getPods(l, deployment.Namespace, deployment.Spec.Selector.MatchLabels)
	if err != nil {
		return err
	}
	return validateResource("pod", *pod, fmt.Sprintf("for deployment %q", deployment.Name), available)
}

func ValidateContainer(l *logrus.Entry, deployment *v1.Deployment, container *string) error {
	if *container == "" {
		*container = deployment.Name
	}
	available := getContainers(l, deployment)
	return validateResource("container", *container, fmt.Sprintf("for deployment %q", deployment.Name), available)
}

func ValidateImage(l *logrus.Entry, deployment *v1.Deployment, container string, image, tag *string) error {
	if *image == "" {
		for _, c := range deployment.Spec.Template.Spec.Containers {
			if container == c.Name {
				pieces := strings.Split(c.Image, ":")
				if len(pieces) != 2 {
					return fmt.Errorf("deployment image %q has invalid format", c.Image)
				}
				*image = pieces[0]
				*tag = pieces[1]
				return nil
			}
		}
	}
	return nil
}

func ValidatePath(wd string, p *string) error {
	if !filepath.IsAbs(*p) {
		*p = path.Join(wd, *p)
	}
	absPath, err := filepath.Abs(*p)
	if err != nil {
		return err
	}
	_, err = os.Stat(absPath)
	if err != nil {
		return err
	}
	*p = absPath
	return nil
}

func ValidateMounts(wd string, ms []string) ([]Mount, error) {
	var mounts []Mount
	for _, m := range ms {
		pieces := strings.Split(m, ":")
		if len(pieces) != 2 {
			return nil, fmt.Errorf("bad format for mount %q, should be %q separated", m, ":")
		}
		hostPath := pieces[0]
		mountPath := pieces[1]
		if err := ValidatePath(wd, &hostPath); err != nil {
			return nil, fmt.Errorf("bad format for mount %q, host path bad: %s", m, err)
		}
		if !path.IsAbs(mountPath) {
			return nil, fmt.Errorf("bad format for mount %q, mount path should be absolute", m)
		}
		mounts = append(mounts, Mount{hostPath, mountPath})
	}
	return mounts, nil

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

func buildDummy(l *logrus.Entry, image, tag, path string) (string, error) {
	cmd := []string{
		"docker", "build", ".",
		"--build-arg", fmt.Sprintf("IMAGE=%v:%v", image, tag),
		"-t", "dummy:latest",
	}
	return command(l, cmd...).cwd(path).run()
}

func debugBuild(l *logrus.Entry, input, goModDir, output string, env []string) (string, error) {
	cmd := []string{
		"go", "build",
		`-gcflags="all=-N -l"`,
		"-o", output, input,
	}
	return command(l, cmd...).cwd(goModDir).env(env).run()
}

func getArgsFromPod(l *logrus.Entry, namespace, pod, container string) ([]string, error) {
	out, err := execPod(l, pod, container, namespace, []string{"cat", "/args.yaml"}).run()
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
	_, err := copyToPod(l, pod, container, deployment.Namespace, argsSource, argsDestination).run()
	if err != nil {
		return err
	}
	return nil
}

func signalCapture(l *logrus.Entry) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, os.Interrupt)
	go func() {
		l.Warnf("signal %s recieved", <-sigchan)
	}()
}

func checkDelveServer(l *logrus.Entry, host string, port int, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%v:%v", host, port), timeout)
	if err != nil {
		return err
	}
	defer conn.Close()
	timer := time.AfterFunc(timeout, func() {
		//rpc2 client tends to hang if it cannot connect
		conn.Close()
	})
	client := rpc2.NewClientFromConn(conn)
	defer client.Disconnect(false)
	if !timer.Stop() {
		return fmt.Errorf("delve connection timed out after %s", timeout)
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
	return command(l, cmd...).run()
}

func tryDelveServer(l *logrus.Entry, host string, port, tries int, sleep time.Duration) error {
	err := tryCall(tries, func(i int) error {
		l.Infof("checking delve connection on %v:%v (%v/%v)", host, port, i, tries)
		return checkDelveServer(l, host, port, 1*time.Second)
	})
	if err != nil {
		return err
	}
	l.Infof("delve server listening on %v:%v", host, port)
	return nil
}

func launchVscode(l *logrus.Entry, goModDir, pod, host string, port, tries int, sleep time.Duration) error {
	command(l, "code", goModDir).postEnd(func() error {
		return tryCall(tries, func(i int) error {
			l.Infof("waiting for vscode status (%v/%v)", i, tries)
			_, err := command(l, "code", "-s").run()
			return err
		})
	}).run()

	l.Infof("opening debug configuration")
	la, err := newLaunchArgs(pod, host, port).toJson()
	if err != nil {
		return err
	}
	_, err = runOpen(l, `vscode://fabiospampinato.vscode-debug-launcher/launch?args=`+la)
	if err != nil {
		return err
	}
	return nil
}

func tryCall(tries int, f func(i int) error) error {
	var err error
	for i := 1; i < tries+1; i++ {
		err = f(i)
		if err == nil {
			return nil
		}
	}
	return err
}

func findGoProjectRoot(path string) (string, error) {
	abs, errAbs := filepath.Abs(path)
	if errAbs != nil {
		return "", errAbs
	}
	dir := filepath.Dir(abs)
	statDir, errStatDir := os.Stat(dir)
	if errStatDir != nil {
		return "", errStatDir
	}
	if !statDir.IsDir() {
		return "", fmt.Errorf("%q is not a dir", dir)
	}
	modFile := filepath.Join(dir, "go.mod")
	stat, errStat := os.Stat(modFile)
	if errStat == nil {
		if stat.IsDir() {
			return "", fmt.Errorf("go.mod is a directory")
		}
		return dir, nil
	}
	if dir == "/" {
		return "", fmt.Errorf("reached / without finding go.mod")
	}
	return findGoProjectRoot(dir)
}
