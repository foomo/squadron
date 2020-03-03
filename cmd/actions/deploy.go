package actions

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVarP(&flagBuild, "build", "", false, "Build deployment before publishing")
}

var (
	flagBuild bool
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy [NAMESPACE] [DEPLOYMENT] -t {TAG}",
		Short: "Deploys the specified deployment with given tag version",
		Long:  `Deploys the specified deployment with given tag version`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			Deploy(args[0], args[1], FlagTag)
		},
	}
)

func Deploy(namespace, deployment, tag string) {
	sds := cnf.GetServiceDeployments(namespace, deployment)
	if len(sds) == 0 {
		log.Fatalf("could not find any service deployments for namespace: %v and deployment: %v", namespace, deployment)
	}
	if flagBuild {
		for _, serviceDeployment := range sds {
			Build(serviceDeployment.ServiceName, tag)
		}
	}
	err := cnf.Deploy(sds, cwdir)
	if err != nil {
		log.Fatal(err)
	}
}
