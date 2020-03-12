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
		_, err := Build(args[0], flagTag, flagDir, flagVerbose)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Build(service, tag, dir string, flagVerbose bool) (string, error) {
	svc, err := cnf.Service(service)
	if err != nil {
		return "", fmt.Errorf("service not found: %v", service)
	}
	output, err := svc.RunBuild(log, dir, tag, flagVerbose)
	if err != nil {
		return "", fmt.Errorf("could not build service: %v output:\n%v", svc.Name, output)
	}
	return output, nil
}
