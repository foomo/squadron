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
			if cmd.Name() == "help" || cmd.Name() == "init" || cmd.Name() == "version" {
				return nil
			}
			// flagDir
			if err := util.ValidatePath(".", &flagDir); err != nil {
				return err
			}
			// cnf
			var err error
			sq, err = squadron.New(log, flagTag, flagDir, flagNamespace)
			if err != nil {
				return err
			}
			return nil
		},
	}

	log           *logrus.Entry
	sq            *squadron.Squadron
	flagTag       string
	flagDir       string
	flagVerbose   bool
	flagNamespace string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagNamespace, "namespace", "n", "default", "Specifies the namespace")
	rootCmd.PersistentFlags().StringVarP(&flagTag, "tag", "t", "latest", "Specifies the image tag")
	rootCmd.PersistentFlags().StringVarP(&flagDir, "dir", "d", "", "Specifies working directory")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Specifies should command output be displayed")
	rootCmd.AddCommand(buildCmd, installCmd, genCmd, uninstallCmd, restartCmd, initCmd, versionCmd)
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
