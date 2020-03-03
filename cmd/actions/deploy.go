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
	sds, err := cnf.ServiceDeployments(cwdir, namespace, deployment)
	if err != nil {
		log.Fatalf("could not deploy: %v  output:\n%v", deployment, err)
	}
	if flagBuild {
		for _, serviceDeployment := range sds {
			Build(serviceDeployment.ServiceName, FlagTag)
		}
	}
	for _, serviceDeployment := range sds {
		cnf.Deploy(serviceDeployment, namespace, tag)
	}
}
