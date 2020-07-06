package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service squadron before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the built service to the registry")
	installCmd.Flags().StringSliceVar(&flagTemplateSlice, "template-vars", nil, "Specifies template vars x=y")
	installCmd.Flags().StringVar(&flagTemplateFile, "template-file", "", "Specifies the template file with vars")
	installCmd.Flags().BoolVar(&flagChartApiV1, "chart-api-v1", false, "Use chart API v1 when creating chart")
}

var (
	flagBuild         bool
	flagOutputDir     string
	flagTemplateSlice []string
	flagTemplateFile  string
	flagChartApiV1    bool
)

var (
	installCmd = &cobra.Command{
		Use:   "install [SQUADRON] -n {NAMESPACE} -t {TAG}",
		Short: "installs a squadron of services",
		Long:  "installs a squadron of services with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateVars, err := squadron.NewTemplateVars(flagDir, flagTemplateSlice, flagTemplateFile)
			if err != nil {
				return err
			}
			_, err = install(args[0], flagNamespace, flagOutputDir, flagBuild, templateVars, flagChartApiV1)
			return err
		},
	}
)

func install(group, namespace, outputDir string, buildService bool, tv squadron.TemplateVars, useChartApiV1 bool) (string, error) {
	ns, err := sq.Namespace(namespace)
	if err != nil {
		return "", err
	}
	g, err := ns.Group(group)
	if err != nil {
		return "", err
	}
	services := g.Services()

	if buildService {
		log.Infof("Building services")
		for _, service := range services {
			out, err := build(service, flagPush)
			if err != nil {
				if err == squadron.ErrBuildNotConfigured {
					log.Warnf("Build command not set for service %q, skipping", service)
				} else {
					return out, err
				}
			}
		}
	}
	if err := sq.CheckIngressController("ingress-nginx"); err != nil {
		return "", err
	}
	return sq.Install(namespace, g.Name, g.Version, services, tv, outputDir, useChartApiV1)
}
