package actions

import (
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
			if flagBuild {
				//TODO: Get All Services From Deployment and build them
			}
			Deploy(args[0], args[1], FlagTag)
		},
	}
)

func Deploy(namespace, deployment, tag string) {
	cnf.Deploy(namespace, deployment, tag)
}
