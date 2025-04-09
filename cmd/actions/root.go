package actions

import (
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	cowsay "github.com/Code-Hex/Neo-cowsay/v2"
	"github.com/foomo/squadron/internal/cmd"
	"github.com/foomo/squadron/internal/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cwd  string
	root *cobra.Command
)

func init() {
	root = NewRoot()
	root.AddCommand(
		NewUp(NewViper(root)),
		NewDiff(NewViper(root)),
		NewDown(NewViper(root)),
		NewBuild(NewViper(root)),
		NewPush(NewViper(root)),
		NewList(NewViper(root)),
		NewRollback(NewViper(root)),
		NewStatus(NewViper(root)),
		NewConfig(NewViper(root)),
		NewVersion(NewViper(root)),
		NewCompletion(NewViper(root)),
		NewTemplate(NewViper(root)),
		NewPostRenderer(NewViper(root)),
		NewSchema(NewViper(root)),
	)
}

// NewRoot represents the base command when called without any subcommands
func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:           "squadron",
		Short:         "Docker compose for kubernetes",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if viper.GetBool("debug") {
				pterm.EnableDebugMessages()
			}
			if cmd.Name() == "help" || cmd.Name() == "init" || cmd.Name() == "version" {
				return nil
			}
			return util.ValidatePath(".", &cwd)
		},
	}

	flags := root.PersistentFlags()
	flags.BoolP("debug", "d", false, "show all output")
	_ = viper.BindPFlag("debug", root.PersistentFlags().Lookup("debug"))

	flags.StringSliceP("file", "f", []string{"squadron.yaml"}, "specify alternative squadron files")

	return root
}

func NewViper(root *cobra.Command) *viper.Viper {
	c := viper.New()
	_ = c.BindPFlag("file", root.PersistentFlags().Lookup("file"))
	return c
}

func Execute() {
	l := cmd.NewLogger()

	say := func(msg string) string {
		if say, cerr := cowsay.Say(msg, cowsay.BallonWidth(80)); cerr == nil {
			msg = say
		}
		return msg
	}

	code := 0
	defer func() {
		if r := recover(); r != nil {
			l.Error(say("It's time to panic"))
			l.Error(fmt.Sprintf("%v", r))
			l.Error(string(debug.Stack()))
			code = 1
		}
		os.Exit(code)
	}()

	if err := root.Execute(); err != nil {
		l.Error(util.SprintError(err))
		code = 1
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
