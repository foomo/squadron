package actions

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/foomo/squadron"
	"github.com/foomo/squadron/util"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use: "squadron",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log = newLogger(flagVerbose)
			var err error
			if cmd.Name() == "help" || cmd.Name() == "init" || cmd.Name() == "version" {
				return nil
			}
			// cwd
			if err = util.ValidatePath(".", &cwd); err != nil {
				return err
			}
			return nil
		},
	}

	log           *logrus.Entry
	cwd           string
	flagVerbose   bool
	flagNamespace string
	flagBuild     bool
	flagPush      bool
	flagFiles     []string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "show more output")
	rootCmd.PersistentFlags().StringSliceVarP(&flagFiles, "file", "f", []string{"squadron.yaml"}, "specify alternative squadron files")

	rootCmd.AddCommand(upCmd, downCmd, buildCmd, generateCmd, configCmd, versionCmd, completionCmd)
}

func Execute() {
	log := logrus.New()
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func newLogger(verbose bool) *logrus.Entry {
	logger := logrus.New()
	if verbose {
		logger.SetLevel(logrus.TraceLevel)
	}
	return logrus.NewEntry(logger)
}

// parseExtraArgs ...
func parseExtraArgs(args []string) (out []string, extraArgs []string) {
	for i, arg := range args {
		if arg == "--" {
			return args[:i], args[i+1:]
		} else if strings.HasPrefix(arg, "--") && i > 0 {
			return args[:i-1], args[i:]
		} else if strings.HasPrefix(arg, "--") {
			return nil, args
		}
	}
	return
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
