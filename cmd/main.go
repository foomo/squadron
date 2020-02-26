package main

import (
	"github.com/spf13/cobra"
)

func main() {
	var cmdAdd = &cobra.Command{
		Use:   "add [configuration file to use for service group]",
		Short: "Add a service group",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			Add(args[0])
		},
	}
	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(cmdAdd)
	rootCmd.Execute()
}
