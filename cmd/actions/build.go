package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cmdBuild = &cobra.Command{
	Use:   "build [configuration file to use for service] -t TAG",
	Short: "Build a service",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := Build(args[0], FlagTag, FlagDir)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(cmdBuild)
}

func Build(service, tag, dir string) (string, error) {
	svc, err := cnf.Service(service)
	if err != nil {
		return "", fmt.Errorf("service not found: %v", err)
	}
	output, err := svc.RunBuild(log, dir, tag)
	if err != nil {
		return "", fmt.Errorf("could not build service: %v output:\n%v", svc.Name, output)
	}
	return output, nil
}
