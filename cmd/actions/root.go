package actions

import (
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

func parseExtraArgs(args []string) (out []string, extraArgs []string) {
	for i, arg := range args {
		if arg == "--" {
			out, extraArgs = args[:i], args[i+1:]
			break
		}
	}
	return
}
