package actions

import (
	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service group before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().StringVarP(&flagService, "service", "s", "", "Specifies the service to work with")
}

var (
	flagBuild     bool
	flagOutputDir string
	flagService   string
)

var (
	installCmd = &cobra.Command{
		Use:   "install [GROUP] -n {NAMESPACE} -t {TAG} -s {SERVICE}",
		Short: "installs a group of services",
		Long:  "installs a group of services with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log := newLogger(flagVerbose)
			return configurd.CheckIngressController(log, "ingress-nginx")
		},
		Run: func(cmd *cobra.Command, args []string) {
			_, err := install(args[0], flagNamespace, flagTag, flagDir, flagOutputDir, flagService, flagBuild, flagVerbose)
			if err != nil {
				log.WithError(err).Fatalf("Installation failed")
			}
		},
	}
)

func install(group, namespace, tag, workDir, outputDir, service string, buildService, verbose bool) (string, error) {
	log := newLogger(verbose)
	cnf := mustNewConfigurd(log, tag, workDir)

	ns, err := cnf.Namespace(namespace)
	if err != nil {
		return "", err
	}
	g, err := ns.Group(group)
	if err != nil {
		return "", err
	}
	sis, err := g.ServiceItems()
	if err != nil {
		return "", err
	}

	if service != "" {
		si, err := g.ServiceItem(service)
		if err != nil {
			return "", err
		}
		sis = []configurd.ServiceItem{si}
	}

	if buildService {
		log.Printf("Building services")
		for _, si := range sis {
			output, err := build(si.Name, tag, workDir, true, verbose)
			if err != nil {
				return output, err
			}
		}
	}
	return cnf.Install(configurd.InstallConfiguration{
		ServiceItems: sis,
		BasePath:     workDir,
		OutputDir:    outputDir,
		Tag:          tag,
		Verbose:      verbose,
	})
}
