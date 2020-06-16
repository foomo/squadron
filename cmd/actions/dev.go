package actions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	devCmd.PersistentFlags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	devCmd.PersistentFlags().StringVarP(&flagContainer, "container", "c", "", "container name to be patched, using deployment name by default")
	devPatchCmd.Flags().StringVarP(&flagImage, "image", "i", "", "Image to be used for dev pod")
	devPatchCmd.Flags().StringVarP(&flagMount, "mount", "m", ".", "host path to be mounted to dev pod")
	devPatchCmd.Flags().BoolVar(&flagRollback, "rollback", false, "rollback deployment to a previous state")
	devDelveCmd.Flags().StringVar(&flagInput, "input", ".", "go file input")
	devDelveCmd.Flags().StringVar(&flagArgs, "args", "", "go file args")
	devDelveCmd.Flags().BoolVar(&flagCleanup, "cleanup", false, "cleanup delve debug session")
	devCmd.AddCommand(devPatchCmd, devShellCmd, devDelveCmd)
}

var (
	devCmd      = &cobra.Command{Use: "dev"}
	devPatchCmd = &cobra.Command{
		Use:   "patch [DEPLOYMENT] -c {CONTAINER} -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "applies a development patch for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if flagContainer == "" {
				flagContainer = args[0]
			}
			_, err := patch(args[0], flagContainer, flagNamespace, flagImage, flagTag, flagMount, flagRollback, flagVerbose)
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
			_, err := shell(args[0], flagNamespace, flagVerbose)
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
			if flagContainer == "" {
				flagContainer = args[0]
			}
			var cArgs []string
			if flagArgs != "" {
				cArgs = strings.Split(flagArgs, " ")
			}
			_, err := delve(args[0], flagInput, flagNamespace, flagContainer, cArgs, flagCleanup, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("debug in dev mode failed")
			}
		},
	}
	flagImage     string
	flagMount     string
	flagContainer string
	flagInput     string
	flagArgs      string
	flagCleanup   bool
	flagRollback  bool
)

func patch(deployment, container, namespace, image, tag, mountPath string, rollback, verbose bool) (string, error) {
	log := newLogger(verbose)
	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(absMountPath)
	if err != nil {
		return "", err
	}
	isPatched, err := configurd.DeploymentIsPatched(log, deployment, namespace)
	if err != nil {
		return "", err
	}

	if rollback {
		return configurd.Rollback(log, deployment, namespace, isPatched)
	}
	return configurd.Patch(log, deployment, container, namespace, image, tag, absMountPath, isPatched)
}

func shell(deployment, namespace string, verbose bool) (string, error) {
	log := newLogger(verbose)
	isPatched, err := configurd.DeploymentIsPatched(log, deployment, namespace)
	if err != nil {
		return "", err
	}

	return configurd.ShellDev(log, deployment, namespace, isPatched)
}

func delve(deployment, input, namespace, container string, args []string, cleanup, verbose bool) (string, error) {
	log := newLogger(verbose)
	isPatched, err := configurd.DeploymentIsPatched(log, deployment, namespace)
	if err != nil {
		return "", err
	}

	if cleanup {
		return configurd.DelveCleanup(log, namespace, deployment, container, isPatched)
	}
	return configurd.Delve(log, namespace, deployment, container, input, args, isPatched)
}
