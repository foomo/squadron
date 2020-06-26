package actions

import (
	"github.com/foomo/squadron"
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

			var err error
			// flagDir
			if err := ValidatePath(".", &flagDir); err != nil {
				return err
			}
			// templateVars
			templateVars, err = squadron.NewTemplateVars(flagDir, flagTemplateSlice, flagTemplateFile)
			if err != nil {
				return err
			}
			// cnf
			sq, err = squadron.New(log, flagTag, flagDir)
			if err != nil {
				return err
			}
			return nil
		},
	}

	log               *logrus.Entry
	sq                *squadron.Squadron
	templateVars      squadron.TemplateVars
	flagTag           string
	flagDir           string
	flagVerbose       bool
	flagNamespace     string
	flagTemplateSlice []string
	flagTemplateFile  string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagTag, "tag", "t", "latest", "Specifies the image tag")
	rootCmd.PersistentFlags().StringVarP(&flagDir, "dir", "d", "", "Specifies working directory")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Specifies should command output be displayed")
	rootCmd.PersistentFlags().StringSliceVar(&flagTemplateSlice, "template-vars", nil, "Specifies template vars x=y")
	rootCmd.PersistentFlags().StringVar(&flagTemplateFile, "template-file", "", "Specifies the template file with vars")
	rootCmd.AddCommand(buildCmd, installCmd, uninstallCmd, initCmd, versionCmd)
}

func Execute() {
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
