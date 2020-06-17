package actions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	devCmd.PersistentFlags().StringVarP(&flagNamespace, "namespace", "n", "default", "namespace name")
	devCmd.PersistentFlags().StringVarP(&flagContainer, "container", "c", "", "container name, using deployment name by default")
	devPatchCmd.Flags().StringVarP(&flagImage, "image", "i", "", "Image to be used for dev pod")
	devPatchCmd.Flags().StringVarP(&flagMount, "mount", "m", ".", "host path to be mounted to dev pod")
	devPatchCmd.Flags().BoolVar(&flagRollback, "rollback", false, "rollback deployment to a previous state")
	devDelveCmd.Flags().StringVar(&flagInput, "input", ".", "go file input")
	devDelveCmd.Flags().StringVar(&flagArgs, "args", "", "go file args")
	devDelveCmd.Flags().BoolVar(&flagCleanup, "cleanup", false, "cleanup delve debug session")
	devDelveCmd.Flags().BoolVar(&flagDlvContinue, "continue", false, "delve --continue option")
	devDelveCmd.Flags().IntVar(&flagDlvPort, "port", 0, "delve port to listen on")
	devDelveCmd.Flags().StringVarP(&flagPod, "pod", "p", "", "pod name, using most recent one by default")
	devCmd.AddCommand(devPatchCmd, devShellCmd, devDelveCmd)
}

var (
	devCmd      = &cobra.Command{Use: "dev"}
	devPatchCmd = &cobra.Command{
		Use:   "patch [DEPLOYMENT] -c {CONTAINER} -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "applies a development patch for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := patch(flagNamespace, args[0], flagPod, flagContainer, flagImage, flagTag, flagMount, flagRollback, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("dev mode failed")
			}
		},
	}
	devShellCmd = &cobra.Command{
		Use:   "shell [DEPLOYMENT] -n {NAMESPACE} -c {CONTAINER}",
		Short: "shell into the dev patched deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := shell(flagNamespace, args[0], flagPod, flagContainer, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("shelling into dev mode failed")
			}
		},
	}
	devDelveCmd = &cobra.Command{
		Use:   "delve [DEPLOYMENT] -input {INPUT} -n {NAMESPACE} -c {CONTAINER}",
		Short: "start a headless delve debug server for .go input on a patched deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := delve(flagNamespace, args[0], flagPod, flagContainer, flagInput, flagArgs, flagDlvContinue, flagDlvPort, flagCleanup, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("debug in dev mode failed")
			}
		},
	}
	flagImage       string
	flagMount       string
	flagPod         string
	flagContainer   string
	flagInput       string
	flagArgs        string
	flagCleanup     bool
	flagRollback    bool
	flagDlvContinue bool
	flagDlvPort     int
)

func patch(namespace, deployment, pod, container, image, tag, mountPath string, rollback, verbose bool) (string, error) {
	log := newLogger(verbose)

	if container == "" {
		container = deployment
	}
	d, err := configurd.ValidateResources(log, namespace, deployment, pod, container)
	if err != nil {
		return "", err
	}

	if image == "" {
		log.Infof("getting image and tag from deployment %v", deployment)
		image, tag, err = configurd.GetDeploymentImageTag(d, container)
		if err != nil {
			return "", err
		}
	}

	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(absMountPath)
	if err != nil {
		return "", err
	}

	if rollback {
		return configurd.Rollback(log, d)
	}
	return configurd.Patch(log, d, container, image, tag, absMountPath)
}

func shell(namespace, deployment, pod, container string, verbose bool) (string, error) {
	log := newLogger(verbose)

	if container == "" {
		container = deployment
	}

	d, err := configurd.ValidateResources(log, namespace, deployment, pod, container)
	if err != nil {
		return "", err
	}

	if pod == "" {
		log.Infof("getting most recent pod with selector from deployment %v", deployment)
		pod, err = configurd.GetMostRecentPodBySelectors(log, d.Spec.Selector.MatchLabels, d.Namespace)
		if err != nil {
			return "", err
		}
	}
	return configurd.Shell(log, d, pod)
}

func delve(namespace, deployment, pod, container, input string, flagArgs string, delveContinue bool, delvePort int, cleanup, verbose bool) (string, error) {
	log := newLogger(verbose)

	var args []string
	if flagArgs != "" {
		args = strings.Split(flagArgs, " ")
	}

	if container == "" {
		container = deployment
	}

	var err error
	if delvePort == 0 {
		delvePort, err = configurd.GetFreePort("localhost")
		if err != nil {
			return "", err
		}
	}

	d, err := configurd.ValidateResources(log, namespace, deployment, pod, container)
	if err != nil {
		return "", err
	}

	if pod == "" {
		log.Infof("getting most recent pod with selector from deployment %v", deployment)
		pod, err = configurd.GetMostRecentPodBySelectors(log, d.Spec.Selector.MatchLabels, d.Namespace)
		if err != nil {
			return "", err
		}
	}

	if cleanup {
		return configurd.DelveCleanup(log, d, pod, container)
	}
	return configurd.Delve(log, d, pod, container, input, args, delveContinue, delvePort)
}
