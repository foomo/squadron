package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	installCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service squadron before publishing")
	installCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	installCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the built service to the registry")
	installCmd.Flags().StringSliceVar(&flagTemplateDataSlice, "template-data", nil, "Specifies template data x=y")
	installCmd.Flags().StringVar(&flagTemplateDataFile, "template-data-file", "", "Specifies the template data file")
}

var (
	flagBuild             bool
	flagOutputDir         string
	flagTemplateDataSlice []string
	flagTemplateDataFile  string
)

var (
	installCmd = &cobra.Command{
		Use:   "install [SQUADRON] -n {NAMESPACE} -t {TAG}",
		Short: "installs a squadron of services",
		Long:  "installs a squadron of services with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateVars, err := squadron.NewTemplateVars(flagDir, flagTemplateDataSlice, flagTemplateDataFile)
			if err != nil {
				return err
			}
			_, err = install(args[0], flagNamespace, flagOutputDir, flagBuild, templateVars)
			return err
		},
	}
)

func install(group, namespace, outputDir string, buildService bool, tv squadron.TemplateVars) (string, error) {
	g, err := sq.Group(namespace, group, tv)
	if err != nil {
		return "", err
	}

	if buildService {
		log.Infof("Building services")
		for _, service := range g.Services() {
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
	return sq.Install(g, tv, outputDir)
}
