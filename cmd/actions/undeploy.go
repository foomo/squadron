package actions

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(undeployCmd)
	undeployCmd.Flags().StringVarP(&flagDeployment, "deployment", "d", "", "Specifies the deployment name")
}

var (
	flagDeployment string
)

var (
	undeployCmd = &cobra.Command{
		Use:   "undeploy [NAMESPACE] -d {DEPLOYMENT}",
		Short: "Uneploys the specified deployment with given tag version",
		Long:  `Uneploys the specified deployment with given tag version`,
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			Undeploy(args[0], flagDeployment)
		},
	}
)

func Undeploy(namespace, deployment string) {
	sds := cnf.GetServiceDeployments(namespace, deployment)
	if len(sds) == 0 {
		log.Fatalf("could not find any service deployments for namespace: %v and deployment: %v", namespace, deployment)
	}
	err := cnf.Undeploy(sds)
	if err != nil {
		log.Fatal(err)
	}
}
