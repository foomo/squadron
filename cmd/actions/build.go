package actions

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	buildCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the service to the registry")
}

var (
	flagPush bool
)

var buildCmd = &cobra.Command{
	Use:   "build [SERVICE]",
	Short: "Build a service with a given tag",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		out, err := build(args[0], flagTag, flagDir, flagPush, flagVerbose)
		if err != nil {
			log.WithError(err).WithField("output", out).Fatal("Build failed")
		}
	},
}

func build(service, tag, dir string, push, verbose bool) (string, error) {
	logger := newLogger(verbose)
	cnf := mustNewConfigurd(logger, tag, dir)
	svc, err := cnf.Service(service)
	if err != nil {
		return "", fmt.Errorf("could not find service: %w", err)
	}

	out, err := cnf.Build(svc)
	if err != nil {
		return out, err
	}

	if push {
		return cnf.Push(service)
	}

	return out, nil
}
