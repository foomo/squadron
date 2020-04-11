package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [SERVICE] -t {TAG}",
	Short: "Build a service with a given tag",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := Build(args[0], flagTag, flagVerbose)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Build(service, dir string, flagVerbose bool) (string, error) {
	svc, err := cnf.Service(service)
	if err != nil {
		return "", fmt.Errorf("service not found: %v", service)
	}
	output, err := svc.RunBuild(log, dir, flagVerbose)
	if err != nil {
		return output, outputErrorf(output, err, "could not build service: %v", svc.Name)
	}
	return output, nil
}
