package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVarP(&flagBuild, "build", "", false, "Build service group before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
}

var (
	flagBuild     bool
	flagOutputDir string
)

var (
	installCmd = &cobra.Command{
		Use:   "install [NAMESPACE] [SERVICE GROUP] -t {TAG}",
		Short: "installs the specified service group with given tag version",
		Long:  "installs the specified service group with given tag version",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := install(args[0], args[1], FlagTag, FlagDir, flagBuild)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func install(namespace, serviceGroup, tag, dir string, shouldBuild bool) (string, error) {
	sgis := cnf.GetServiceGroupItems(namespace, serviceGroup)
	if len(sgis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and service group: %v", namespace, serviceGroup)
	}
	if shouldBuild {
		for _, sgi := range sgis {
			_, err := Build(sgi.ServiceName, tag, dir)
			if err != nil {
				return "", err
			}
		}
	}
	output, err := cnf.Install(log, sgis, flagOutputDir, tag)
	if err != nil {
		return "", fmt.Errorf("could not install service group: %v output:\n%v", serviceGroup, output)
	}
	return output, nil
}
