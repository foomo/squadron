package actions

import (
	"fmt"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service group before publishing")
	installCmd.Flags().BoolVarP(&flagUpgrade, "upgrade", "u", false, "Upgrade the service if already installed")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	flagBuild     bool
	flagUpgrade   bool
	flagOutputDir string
)

var (
	installCmd = &cobra.Command{
		Use:   "install [GROUP] -n {NAMESPACE} -t {TAG}",
		Short: "installs a group of services",
		Long:  "installs a group of services with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := install(args[0], flagNamespace, flagTag, flagDir, flagOutputDir, flagBuild, flagUpgrade, flagVerbose)
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func install(group, namespace, tag, workDir, outputDir string, buildService, upgrade, verbose bool) (string, error) {
	cnf := mustNewConfigurd(tag, workDir)
	sis := cnf.GetServiceItems(namespace, group)
	if len(sis) == 0 {
		return "", fmt.Errorf("could not find any service for namespace: %v and group: %v", namespace, group)
	}
	if buildService {
		log.Printf("Building services")
		for _, si := range sis {
			_, err := build(si.Name, tag, workDir, verbose)
			if err != nil {
				return "", err
			}
		}
	}
	output, err := cnf.Install(log, configurd.InstallConfiguration{
		ServiceItems: sis,
		BasePath:     workDir,
		OutputDir:    outputDir,
		Tag:          tag,
		Upgrade:      upgrade,
		Verbose:      verbose,
	})

	if err != nil {
		return output, outputErrorf(output, err, "could not install group: %v", group)
	}
	return output, nil
}
