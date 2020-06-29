package actions

import (
	"github.com/spf13/cobra"
)

func init() {
	restartCmd.Flags().StringVarP(&flagService, "service", "s", "", "Specifies the service to work with")
}

var flagService string

var restartCmd = &cobra.Command{
	Use:   "restart [GROUP] -n {NAMESPACE} -s {SERVICE}",
	Short: "restart a deployment or a service",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := restart(args[0], flagNamespace, flagService)
		return err
	},
}

func restart(group, namespace, service string) (string, error) {
	ns, err := sq.Namespace(namespace)
	if err != nil {
		return "", err
	}
	g, err := ns.Group(group)
	if err != nil {
		return "", err
	}
	services := g.Services()

	if service != "" {
		services = []string{service}
	}

	return sq.Restart(services)
}
