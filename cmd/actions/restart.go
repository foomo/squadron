package actions

import (
	"github.com/foomo/squadron"
	"github.com/spf13/cobra"
)

func init() {
	restartCmd.Flags().StringVarP(&flagService, "service", "s", "", "Specifies the service to work with")
	restartCmd.Flags().StringSliceVar(&flagTemplateDataSlice, "template-data", nil, "Specifies template data x=y")
	restartCmd.Flags().StringVar(&flagTemplateDataFile, "template-data-file", "", "Specifies the template data file")
}

var flagService string

var restartCmd = &cobra.Command{
	Use:   "restart [GROUP] -n {NAMESPACE} -s {SERVICE}",
	Short: "restart a deployment or a service",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateVars, err := squadron.NewTemplateVars(flagDir, flagTemplateDataSlice, flagTemplateDataFile)
		if err != nil {
			return err
		}
		_, err = restart(args[0], flagNamespace, flagService, templateVars)
		return err
	},
}

func restart(group, namespace, service string, tv squadron.TemplateVars) (string, error) {
	g, err := sq.Group(namespace, group, tv)
	if err != nil {
		return "", err
	}
	services := g.Services()

	if service != "" {
		services = []string{service}
	}

	return sq.Restart(services)
}
