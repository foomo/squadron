package actions

import (
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [SERVICE] -t {TAG}",
	Short: "Build a service with a given tag",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := build(args[0], flagTag, flagDir, flagVerbose)
		if err != nil {
			log.WithError(err).Fatalf("Build failed")
		}
	},
}

func build(service, tag, dir string, verbose bool) (string, error) {
	logger := newLogger(verbose)
	cnf := mustNewConfigurd(logger, tag, dir)

	return cnf.Build(service)
}
