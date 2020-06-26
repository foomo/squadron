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
	RunE: func(cmd *cobra.Command, args []string) error {
		_, err := build(args[0], flagPush)
		return err
	},
}

func build(service string, push bool) (string, error) {
	svc, err := sq.Service(service)
	if err != nil {
		return "", fmt.Errorf("could not find service: %w", err)
	}

	out, err := sq.Build(svc)
	if err != nil {
		return out, err
	}

	if push {
		return sq.Push(service)
	}

	return out, nil
}
