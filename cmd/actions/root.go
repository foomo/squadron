package actions

import (
	"os"
	"path"

	"github.com/foomo/configurd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use: "configurd",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Name() == "help" || cmd.Name() == "init" {
				return nil
			}
			// flagDir
			wdir, err := os.Getwd()
			if err != nil {
				return err
			}
			if flagDir != "" {
				flagDir = path.Join(wdir, flagDir)
			} else {
				flagDir = wdir
			}
			// templateVars
			templateVars, err = configurd.NewTemplateVars(flagDir, flagTemplateSlice, flagTemplateFile)
			if err != nil {
				return err
			}
			// cnf
			cnf, err = newConfigurd(newLogger(flagVerbose), flagTag, flagDir, templateVars)
			if err != nil {
				return err
			}
			return nil
		},
	}

	cnf               configurd.Configurd
	templateVars      configurd.TemplateVars
	flagTag           string
	flagDir           string
	flagVerbose       bool
	flagNamespace     string
	flagTemplateSlice []string
	flagTemplateFile  string
)

func newConfigurd(log *logrus.Entry, tag, basePath string, tv configurd.TemplateVars) (configurd.Configurd, error) {
	config := configurd.Config{
		Tag:          tag,
		BasePath:     basePath,
		Log:          log,
		TemplateVars: tv,
	}

	return configurd.New(config)
}

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
