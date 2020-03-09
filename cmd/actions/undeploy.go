package actions

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(undeployCmd)
	undeployCmd.Flags().StringVarP(&FlagSG, "sg", "g", "", "Specifies the service group name")
}

var (
	FlagSG string
)

var (
	undeployCmd = &cobra.Command{
		Use:   "undeploy [NAMESPACE] -g {SERVICE GROUP}",
		Short: "Uneploys the specified service group with given tag version",
		Long:  `Uneploys the specified service group with given tag version`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			Undeploy(args[0], FlagSG)
		},
	}
)

func Undeploy(namespace, deployment string) {
	sgis := cnf.GetServiceGroupItems(namespace, deployment)
	if len(sgis) == 0 {
		log.Fatalf("could not find any service for namespace: %v and service group: %v", namespace, deployment)
	}
	err := cnf.Undeploy(sgis)
	if err != nil {
		log.Fatal(err)
	}
}
