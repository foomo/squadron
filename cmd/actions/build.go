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
		_, err := build(args[0], flagTag, flagDir, flagVerbose)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func build(service, tag, dir string, flagVerbose bool) (string, error) {
	cnf := mustNewConfigurd(dir, tag)
	svc, err := cnf.Service(service)
	if err != nil {
		return "", fmt.Errorf("service not found: %v", service)
	}
	output, err := svc.RunBuild(log, dir, tag, flagVerbose)
	if err != nil {
		return output, outputErrorf(output, err, "could not build service: %v", svc.Name)
	}
	return output, nil
}
