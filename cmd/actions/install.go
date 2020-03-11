package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVarP(&flagBuild, "build", "", false, "Build service group before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().StringVarP(&FlagNamespace, "namespace", "n", "default", "Specifies the namespace")
	// installCmd.Flags().StringVarP(&FlagGroup, "group", "g", "", "Specifies the group")
}

var (
	flagBuild     bool
	flagOutputDir string
)

var (
	installCmd = &cobra.Command{
		Use:   "install [GROUP] -n {NAMESPACE} -t {TAG}",
		Short: "installs the specified group with given namespace and tag version",
		Long:  "installs the specified group with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := install(args[0], FlagNamespace, FlagTag, FlagDir, flagOutputDir, flagBuild, FlagVerbose)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func install(group, namespace, tag, workDir, outputDir string, build, verbose bool) (string, error) {
	sis := cnf.GetServiceItems(namespace, group)
	if len(sis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and group: %v", namespace, group)
	}
	if build {
		log.Printf("Building services")
		for _, si := range sis {
			_, err := Build(si.ServiceName, tag, workDir, verbose)
			if err != nil {
				return "", err
			}
		}
	}
	output, err := cnf.Install(log, sis, workDir, outputDir, tag, verbose)
	if err != nil {
		// return "", fmt.Errorf("could not install group: %v output:\n%v \nerror: \n%v", group, output, err)
		return "", errorf(output, err, "could not install group: %v", group)
	}
	return output, nil
}

func errorf(output string, err error, format string, args ...interface{}) error {
	return fmt.Errorf("%v, error: %v", fmt.Sprintf(format, args...), err)
}
