package actions

import (
	"os"
	"path/filepath"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	devCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	devCmd.Flags().StringVarP(&flagImage, "image", "i", "golang", "Image to be used for dev pod")
	devCmd.Flags().StringVarP(&flagMount, "mount", "m", ".", "host path to be mounted to dev pod")
	devCmd.Flags().StringVarP(&flagContainer, "container", "c", "", "container name to be patched, using deployment name by default")
}

var (
	devCmd = &cobra.Command{
		Use:   "dev [DEPLOYMENT] -c {CONTAINER} -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "creates a dev deployment for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := dev(args[0], flagContainer, flagNamespace, flagImage, flagTag, flagMount, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Installation failed")
			}
		},
	}
	flagImage     string
	flagMount     string
	flagContainer string
)

func dev(deployment, container, namespace, image, tag, mountPath string, verbose bool) (string, error) {
	log := newLogger(verbose)

	if container == "" {
		container = deployment
	}
	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(absMountPath)
	if err != nil {
		return "", err
	}
	return configurd.RolloutDev(log, deployment, container, namespace, image, tag, absMountPath)
}
