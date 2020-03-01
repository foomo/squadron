package actions

import (
	"context"
	"log"
	"os"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

var cmdBuild = &cobra.Command{
	Use:   "build [configuration file to use for service group] -t TAG",
	Short: "Add a service group",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		dir, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		Build(dir, args[0], "latest")
	},
}

func init() {
	rootCmd.AddCommand(cmdBuild)
}

func Build(dir, service, tag string) {
	cnf, err := configurd.New(dir)
	if err != nil {
		log.Fatal(err)
	}

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
