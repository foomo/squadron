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
	devStartCmd.Flags().StringVarP(&flagImage, "image", "i", "golang", "Image to be used for dev pod")
	devStartCmd.Flags().StringVarP(&flagMount, "mount", "m", ".", "host path to be mounted to dev pod")
	devStartCmd.Flags().BoolVar(&flagGoDebug, "goDebug", false, "use dev mode for go debugging")
	devDelveCmd.Flags().StringVar(&flagInput, "input", ".", "go file input")
	devDelveCmd.Flags().StringVar(&flagArgs, "args", "", "go file args")
	devDelveCmd.Flags().BoolVar(&flagCleanup, "cleanup", false, "cleanup delve debug session")
	devCmd.AddCommand(devStartCmd, devStopCmd, devShellCmd, devDelveCmd)
}

var (
	devCmd      = &cobra.Command{Use: "dev"}
	devStartCmd = &cobra.Command{
		Use:   "start [DEPLOYMENT] -c {CONTAINER} -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "applies a development patch for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if flagContainer == "" {
				flagContainer = args[0]
			}
			_, err := start(args[0], flagContainer, flagNamespace, flagImage, flagTag, flagMount, flagGoDebug, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Starting dev mode failed")
			}
		},
	}
	devStopCmd = &cobra.Command{
		Use:   "stop [DEPLOYMENT] -n {NAMESPACE}",
		Short: "rolls back the development patch for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := stop(args[0], flagNamespace, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Stopping dev mode failed")
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
				log.WithError(err).Fatalf("Shelling into dev mode failed")
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
				log.WithError(err).Fatalf("Debug in dev mode failed")
			}
		},
	}
	flagImage     string
	flagMount     string
	flagContainer string
	flagGoDebug   bool
	flagInput     string
	flagArgs      string
	flagCleanup   bool
)

func start(deployment, container, namespace, image, tag, mountPath string, goDebug, verbose bool) (string, error) {
	log := newLogger(verbose)

	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(absMountPath)
	if err != nil {
		return "", err
	}
	return configurd.RolloutDev(log, deployment, container, namespace, image, tag, absMountPath, goDebug)
}

func stop(deployment, namespace string, verbose bool) (string, error) {
	log := newLogger(verbose)
	return configurd.RollbackDev(log, deployment, namespace)
}

func shell(deployment, namespace string, verbose bool) (string, error) {
	log := newLogger(verbose)
	return configurd.ShellDev(log, deployment, namespace)
}

func delve(deployment, input, namespace, container string, args []string, cleanup, verbose bool) (string, error) {
	log := newLogger(verbose)
	if cleanup {
		return configurd.DelveDevCleanup(log, namespace, deployment, container)
	}
	return configurd.DelveDev(log, namespace, deployment, container, input, args)
}
