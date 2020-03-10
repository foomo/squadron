package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(uninstallCmd)
	uninstallCmd.Flags().StringVarP(&FlagServiceGroup, "service-group", "g", "", "Specifies the service group name")
}

var (
	FlagServiceGroup string
)

var (
	uninstallCmd = &cobra.Command{
		Use:   "uninstall [NAMESPACE] -g {SERVICE GROUP}",
		Short: "uninstalls the specified service group with given tag version",
		Long:  "uninstalls the specified service group with given tag version",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := uninstall(args[0], FlagServiceGroup)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func uninstall(namespace, serviceGroup string) (string, error) {
	sgis := cnf.GetServiceGroupItems(namespace, serviceGroup)
	if len(sgis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and service group: %v", namespace, serviceGroup)
	}
	output, err := cnf.Uninstall(log, sgis)
	if err != nil {
		return "", fmt.Errorf("could not uninstall service group: %v output:\n%v", serviceGroup, output)
	}
	return output, nil
}
