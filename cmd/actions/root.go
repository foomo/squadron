package actions

import (
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
			// squadron
			sq, err = squadron.New(log, cwd, flagNamespace)
			if err != nil {
				return err
			}
			return nil
		},
	}

	log         *logrus.Entry
	sq          *squadron.Squadron
	cwd         string
	flagVerbose bool
)

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
