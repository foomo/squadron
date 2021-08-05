package actions

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/util"
)

var (
	rootCmd = &cobra.Command{
		Use: "squadron",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			logrus.SetLevel(logrus.InfoLevel)
			if flagVerbose {
				logrus.SetLevel(logrus.TraceLevel)
			}
			if cmd.Name() == "help" || cmd.Name() == "init" || cmd.Name() == "version" {
				return nil
			}
			// cwd
			return util.ValidatePath(".", &cwd)
		},
	}

	cwd           string
	flagVerbose   bool
	flagNoRender  bool
	flagNamespace string
	flagBuild     bool
	flagPush      bool
	flagDiff      bool
	flagFiles     []string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "show more output")
	rootCmd.PersistentFlags().StringSliceVarP(&flagFiles, "file", "f", []string{"squadron.yaml"}, "specify alternative squadron files")

	rootCmd.AddCommand(upCmd, downCmd, buildCmd, pushCmd, listCmd, generateCmd, configCmd, versionCmd, completionCmd, templateCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Fatal(err)
	}
}

// parseExtraArgs ...
func parseExtraArgs(args []string) (out []string, extraArgs []string) {
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") && i > 0 {
			return args[:i], args[i:]
		} else if strings.HasPrefix(arg, "--") {
			return nil, args
		}
	}
	return args, nil
}

// parseUnitArgs helper
func parseUnitArgs(args []string, units map[string]squadron.Unit) (map[string]squadron.Unit, error) {
	if len(args) == 0 {
		return units, nil
	}
	ret := map[string]squadron.Unit{}
	for _, arg := range args {
		if unit, ok := units[arg]; ok {
			ret[arg] = unit
		} else {
			return nil, errors.Errorf("unknown unit name %s", arg)
		}
	}
	return ret, nil
}

func parseUnitNames(args []string, units map[string]squadron.Unit) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
	}
	ret := make([]string, 0, len(args))
	for _, arg := range args {
		if _, ok := units[arg]; ok {
			ret = append(ret, arg)
		} else {
			return nil, errors.Errorf("unknown unit name %s", arg)
		}
	}
	return ret, nil
}
