package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	genCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service squadron before publishing")
	genCmd.Flags().StringVarP(&flagOutputDir, "output", "o", "default", "Specifies output directory")
	genCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the built service to the registry")
	genCmd.Flags().StringSliceVar(&flagTemplateDataSlice, "template-data", nil, "Specifies template data x=y")
	genCmd.Flags().StringVar(&flagTemplateDataFile, "template-data-file", "", "Specifies the template data file")
	genCmd.Flags().BoolVar(&flagChartApiV1, "chart-api-v1", false, "Use helm chart api v1 to generate chart files")
}

var (
	flagChartApiV1 bool
)

var (
	genCmd = &cobra.Command{
		Use:   "gen [SQUADRON] -n {NAMESPACE} -t {TAG}",
		Short: "generate a .tgz helm chart",
		Long:  "generate a .tgz helm chart with given namespace and tag",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			templateVars, err := squadron.NewTemplateVars(flagDir, flagTemplateDataSlice, flagTemplateDataFile)
			if err != nil {
				return err
			}
			_, err = gen(args[0], flagNamespace, flagOutputDir, flagBuild, templateVars, flagChartApiV1)
			return err
		},
	}
)

func gen(group, namespace, outputDir string, buildService bool, tv squadron.TemplateVars, useChartApiV1 bool) (string, error) {
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
	return sq.Generate(g, tv, outputDir, useChartApiV1)
}
