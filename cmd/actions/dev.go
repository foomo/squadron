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
}

var (
	devCmd = &cobra.Command{
		Use:   "dev [DEPLOYMENT] -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "creates a dev deployment for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := dev(args[0], flagNamespace, flagImage, flagTag, flagMount, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Installation failed")
			}
		},
	}
	flagImage string
	flagMount string
)

func dev(deployment, namespace, image, tag, mountPath string, verbose bool) (string, error) {
	log := newLogger(verbose)

	absMountPath, err := filepath.Abs(mountPath)
	if err != nil {
		return "", err
	}
	_, err = os.Stat(absMountPath)
	if err != nil {
		return "", err
	}
	return configurd.RolloutDev(log, deployment, namespace, image, tag, absMountPath)
}
