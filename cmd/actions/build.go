package actions

import (
	"context"
	"log"

	"github.com/spf13/cobra"
)

var cmdBuild = &cobra.Command{
	Use:   "build [configuration file to use for service group] -t TAG",
	Short: "Add a service group",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		Build(args[0], cmd.Flag("tag").Value.String())
	},
}

func init() {
	rootCmd.AddCommand(cmdBuild)
}

func Build(service, tag string) {

	svc, err := cnf.Service(service)
	if err != nil {
		log.Fatalf("service not found: %v", err)
	}
	output, err := svc.Build(context.Background(), tag)
	if err != nil {
		log.Fatalf("could not build: %v  output:\n%v", output, err)
	}
	log.Print(output)
}
