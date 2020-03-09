package actions

import (
	"log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(deployCmd)
	deployCmd.Flags().BoolVarP(&flagBuild, "build", "", false, "Build service group before publishing")
	deployCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
}

var (
	flagBuild     bool
	flagOutputDir string
)

var (
	deployCmd = &cobra.Command{
		Use:   "deploy [NAMESPACE] [SERVICE GROUP] -t {TAG}",
		Short: "Deploys the specified service group with given tag version",
		Long:  `Deploys the specified service group with given tag version`,
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			Deploy(args[0], args[1], FlagTag)
		},
	}
)

func Deploy(namespace, serviceGroup, tag string) {
	sgis := cnf.GetServiceGroupItems(namespace, serviceGroup)
	if len(sgis) == 0 {
		log.Fatalf("could not find any service for namespace: %v and service group: %v", namespace, serviceGroup)
	}
	if flagBuild {
		for _, sgi := range sgis {
			Build(sgi.ServiceName, tag)
		}
	}
	err := cnf.Deploy(sgis, flagOutputDir, tag)
	if err != nil {
		log.Fatal(err)
	}
}
