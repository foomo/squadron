package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	uninstallCmd = &cobra.Command{
		Use:   "uninstall [GROUP] -n {NAMESPACE} -t {TAG}",
		Short: "uninstalls the specified group with given namespace and tag version",
		Long:  "uninstalls the specified group with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := uninstall(args[0], flagNamespace, flagVerbose)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func uninstall(group, namespace string, flagVerbose bool) (string, error) {
	sis := cnf.GetServiceItems(namespace, group)
	if len(sis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and service group: %v", namespace, group)
	}
	output, err := cnf.Uninstall(log, sis, flagVerbose)
	if err != nil {
		return "", fmt.Errorf("could not uninstall service group: %v output:\n%v", group, output)
	}
	return output, nil
}
