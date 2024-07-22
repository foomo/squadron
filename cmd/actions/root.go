package actions

import (
	"os"
	"strings"

	"github.com/foomo/squadron/internal/util"
	"github.com/pterm/pterm"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:           "squadron",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if flagSilent {
				logrus.SetLevel(logrus.ErrorLevel)
			} else if flagDebug {
				logrus.SetLevel(logrus.TraceLevel)
				pterm.EnableDebugMessages()
			} else if flagVerbose {
				logrus.SetLevel(logrus.InfoLevel)
			} else {
				logrus.SetLevel(logrus.WarnLevel)
			}
			if cmd.Name() == "help" || cmd.Name() == "init" || cmd.Name() == "version" {
				return nil
			}
			// cwd
			return util.ValidatePath(".", &cwd)
		},
	}

	cwd           string
	flagSilent    bool
	flagDebug     bool
	flagVerbose   bool
	flagNoRender  bool
	flagNamespace string
	flagRevision  string
	flagBuild     bool
	flagPush      bool
	flagParallel  int
	flagBuildArgs []string
	flagPushArgs  []string
	flagTags      []string
	flagFiles     []string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&flagSilent, "silent", "s", false, "only show errors")
	rootCmd.PersistentFlags().BoolVarP(&flagDebug, "debug", "d", false, "show all output")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "show more output")
	rootCmd.PersistentFlags().StringSliceVarP(&flagFiles, "file", "f", []string{"squadron.yaml"}, "specify alternative squadron files")

	rootCmd.AddCommand(upCmd, diffCmd, downCmd, buildCmd, pushCmd, listCmd, rollbackCmd, statusCmd, configCmd, versionCmd, completionCmd, templateCmd, postRendererCmd)

	pterm.Info = *pterm.Info.WithPrefix(pterm.Prefix{Text: "INFO", Style: pterm.Info.Prefix.Style})
	pterm.Error = *pterm.Info.WithPrefix(pterm.Prefix{Text: "ERROR", Style: pterm.Error.Prefix.Style})
	pterm.Warning = *pterm.Info.WithPrefix(pterm.Prefix{Text: "WARNING", Style: pterm.Warning.Prefix.Style})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		pterm.Error.Println(err.Error())
		os.Exit(1)
	}
}

// parseExtraArgs ...
func parseExtraArgs(args []string) (out []string, extraArgs []string) { //nolint:nonamedreturns
	for i, arg := range args {
		if strings.HasPrefix(arg, "--") && i > 0 {
			return args[:i], args[i:]
		} else if strings.HasPrefix(arg, "--") {
			return nil, args
		}
	}
	return args, nil
}

func parseSquadronAndUnitNames(args []string) (squadron string, units []string) { //nolint:nonamedreturns
	if len(args) == 0 {
		return "", nil
	}
	if len(args) > 0 {
		squadron = args[0]
	}
	if len(args) > 1 {
		units = args[1:]
	}
	return squadron, units
}
