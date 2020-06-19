package actions

import (
	"github.com/foomo/configurd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	log     = logrus.New()
	rootCmd = &cobra.Command{
		Use: "configurd",
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			return configurd.ValidatePath("", &flagDir)
		},
	}
	flagTag       string
	flagDir       string
	flagVerbose   bool
	flagNamespace string
)

func newConfigurd(log *logrus.Entry, tag, basePath string) (configurd.Configurd, error) {
	config := configurd.Config{
		Tag:      tag,
		BasePath: basePath,
		Log:      log,
	}

	return configurd.New(config)
}

func mustNewConfigurd(log *logrus.Entry, tag, basePath string) configurd.Configurd {
	cnf, err := newConfigurd(log, tag, basePath)
	if err != nil {
		log.Fatal(err)
	}
	return cnf
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&flagTag, "tag", "t", "latest", "Specifies the image tag")
	rootCmd.PersistentFlags().StringVarP(&flagDir, "dir", "d", ".", "Specifies working directory")
	rootCmd.PersistentFlags().StringVarP(&flagNamespace, "namespace", "n", "default", "namespace name")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "Specifies should command output be displayed")
	rootCmd.AddCommand(buildCmd, installCmd, uninstallCmd, initCmd, versionCmd, devCmd)
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
