package actions

import (
	"github.com/spf13/cobra"
)

func init() {
	uninstallCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
}

var (
	uninstallCmd = &cobra.Command{
		Use:   "uninstall [GROUP] -n {NAMESPACE} -t {TAG}",
		Short: "uninstalls a group",
		Long:  "uninstalls a group with given namespace and tag version",
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
	_, err = ns.Group(group)
	if err != nil {
		return "", err
	}

	return sq.Uninstall(group, namespace)
}
