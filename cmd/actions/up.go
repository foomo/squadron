package actions

import (
	"github.com/foomo/squadron"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	upCmd.Flags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	upCmd.Flags().StringSliceVarP(&flagFiles, "file", "f", []string{}, "Configuration file to merge")
	upCmd.Flags().BoolVarP(&flagBuild, "build", "b", false, "Build service squadron before publishing")
	upCmd.Flags().BoolVarP(&flagPush, "push", "p", false, "Pushes the built service to the registry")
}

var (
	upCmd = &cobra.Command{
		Use:   "up {UNIT...} -n {NAMESPACE} -b -p -f",
		Short: "builds and installs a group of charts",
		Args:  cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			//todo use -f flag for merging multiple files if provided
			var extraArgs []string
			units := args
			for i, arg := range args {
				if arg == "--" {
					extraArgs = args[i+1:]
					units = args[:i]
					break
				}
			}
			return up(log, units, cwd, flagNamespace, flagBuild, flagPush, flagFiles, extraArgs)
		},
	}
)

func up(l *logrus.Entry, unitNames []string, cwd, namespace string, build, push bool, files []string, extraArgs []string) error {
	sq, err := squadron.New(l, cwd, namespace, files)
	if err != nil {
		return err
	}

	units := map[string]squadron.Unit{}
	if len(unitNames) == 0 {
		units = sq.Units()
	}
	for _, un := range unitNames {
		units[un] = sq.Units()[un]
	}

	for _, unit := range units {
		if build {
			if err := sq.Build(unit); err != nil {
				return err
			}
		}
		if push {
			if err := sq.Push(unit); err != nil {
				return err
			}
		}
	}
	return sq.Up(units, namespace, extraArgs)
}
