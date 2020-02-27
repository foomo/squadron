package main

import (
	"log"
	"os"

	"github.com/foomo/configurd/cmd/actions"
	"github.com/spf13/cobra"
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var cmdBuild = &cobra.Command{
		Use:   "build [configuration file to use for service group] -t TAG",
		Short: "Add a service group",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			actions.Build(dir, args[0], "latest")
		},
	}

	var rootCmd = &cobra.Command{Use: "app"}
	rootCmd.AddCommand(cmdBuild)

	_ = rootCmd.Execute()
}
