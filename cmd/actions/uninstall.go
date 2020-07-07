package actions

import (
	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	uninstallCmd = &cobra.Command{
		Use:   "uninstall [SQUADRON] -n {NAMESPACE} -t {TAG}",
		Short: "uninstalls a squadron",
		Long:  "uninstalls a squadron with given namespace and tag version",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := uninstall(args[0], flagNamespace)
			return err
		},
	}
)

func uninstall(group, namespace string) (string, error) {
	ns, err := sq.Namespace(namespace)
	if err != nil {
		return "", err
	}
	if err = ns.ValidateGroup(group); err != nil {
		return "", err
	}

	return sq.Uninstall(group)
}
