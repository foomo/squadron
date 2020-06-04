package actions

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	devCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	devCmd.Flags().StringVarP(&flagImage, "image", "i", "golang", "Image to be used for dev pod")
	devCmd.Flags().StringVarP(&flagMount, "mount", "m", "", "host path to be mounted to dev pod")
}

var (
	devCmd = &cobra.Command{
		Use:   "dev [SERVICE] -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "creates a dev deployment for a service",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := dev(args[0], flagNamespace, flagDir, flagImage, flagTag, flagMount, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Installation failed")
			}
		},
	}
	flagImage string
	flagMount string
)

func dev(service, namespace, workDir, image, tag, mountPath string, verbose bool) (string, error) {
	log := newLogger(verbose)
	cnf := mustNewConfigurd(log, tag, workDir)
	_, err := cnf.Service(service)
	if err != nil {
		return "", err
	}
	_, err = cnf.Namespace(namespace)
	if err != nil {
		return "", err
	}
	if mountPath == "" {
		mountPath = workDir
	}
	_, err = os.Stat(mountPath)
	if err != nil {
		return "", err
	}
	return cnf.RolloutDev(service, namespace, image, tag, mountPath)
}
