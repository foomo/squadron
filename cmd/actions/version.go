package actions

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "latest"

func NewVersion(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "show version information",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Println(version)
		},
	}

	return cmd
}
