package actions

import (
	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service group before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
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
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := CheckIngressController(log, "ingress-nginx"); err != nil {
				return err
			}
			_, err := install(args[0], flagNamespace, flagTag, flagDir, flagOutputDir, flagService, flagBuild, templateVars)
			return err
		},
	}
)

func install(group, namespace, tag, workDir, outputDir, service string, buildService bool, tv configurd.TemplateVars) (string, error) {
	ns, err := cnf.Namespace(namespace)
	if err != nil {
		return "", err
	}
	g, err := ns.Group(group)
	if err != nil {
		return "", err
	}
	overrides, err := g.Overrides(workDir, namespace, tv)
	if err != nil {
		return "", err
	}

	if service != "" {
		overrides = map[string]configurd.Override{
			service: overrides[service],
		}
	}

	if buildService {
		log.Printf("Building services")
		for name := range overrides {
			output, err := build(name, true)
			if err != nil {
				return output, err
			}
		}
	}
	return cnf.Install(overrides, workDir, outputDir, namespace, group, tag)
}
