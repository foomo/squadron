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
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose ouput")
	rootCmd.AddCommand(upCmd, downCmd, buildCmd, versionCmd)
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
