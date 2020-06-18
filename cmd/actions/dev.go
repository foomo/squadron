package actions

import (
	"github.com/foomo/configurd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func init() {
	devCmd.PersistentFlags().VarP(flagNs, "namespace", "n", "namespace name")
	devCmd.PersistentFlags().VarP(flagContainer, "container", "c", "container name, default deployment name")
	devPatchCmd.Flags().StringVarP(&flagImage, "image", "i", "", "Image to be used for dev pod")
	devPatchCmd.Flags().VarP(flagMount, "mount", "m", "host path to be mounted to dev pod")
	devPatchCmd.Flags().BoolVar(&flagRollback, "rollback", false, "rollback deployment to a previous state")
	devDelveCmd.Flags().Var(flagInput, "input", "go file input")
	devDelveCmd.Flags().Var(flagArgs, "args", "go file args")
	devDelveCmd.Flags().VarP(flagPod, "pod", "p", "pod name, using most recent one by default")
	devDelveCmd.Flags().BoolVar(&flagCleanup, "cleanup", false, "cleanup delve debug session")
	devDelveCmd.Flags().BoolVar(&flagContinue, "continue", false, "delve --continue option")
	devDelveCmd.Flags().Var(flagListen, "listen", "delve host:port to listen on")
	devDelveCmd.Flags().BoolVar(&flagVscode, "vscode", false, "launch a debug configuration in vscode")
	devCmd.AddCommand(devPatchCmd, devShellCmd, devDelveCmd)
}

var (
	devCmd = &cobra.Command{
		Use: "dev",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			l = newLogger(flagVerbose)
			//a custom flag is dependant on deployment name
			return flagDeployment.Set(args[0])
		},
	}
	devPatchCmd = &cobra.Command{
		Use:   "patch [DEPLOYMENT] -c {CONTAINER} -n {NAMESPACE} -i {IMAGE} -t {TAG} -m {MOUNT}",
		Short: "applies a development patch for a deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := patch(l, flagNs, flagDeployment, flagPod, flagContainer, flagImage, flagTag,
				flagMount.String(), flagRollback)
			if err != nil {
				log.WithError(err).Fatalf("dev mode failed")
			}
		},
	}
	devShellCmd = &cobra.Command{
		Use:   "shell [DEPLOYMENT] -n {NAMESPACE} -c {CONTAINER}",
		Short: "shell into the dev patched deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := shell(l, flagDeployment, flagPod)
			if err != nil {
				log.WithError(err).Fatalf("shelling into dev mode failed")
			}
		},
	}
	devDelveCmd = &cobra.Command{
		Use:   "delve [DEPLOYMENT] -input {INPUT} -n {NAMESPACE} -c {CONTAINER}",
		Short: "start a headless delve debug server for .go input on a patched deployment",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			_, err := delve(l, flagDeployment, flagPod, flagContainer, flagInput.String(),
				flagArgs.items, flagListen.Host, flagListen.Port, flagCleanup, flagContinue, flagVscode)
			if err != nil {
				log.WithError(err).Fatalf("debug in dev mode failed")
			}
		},
	}
	l              *logrus.Entry
	flagNs         = newNamespace("default")
	flagDeployment = newDeployment(flagNs)
	flagPod        = newPod(flagDeployment)
	flagContainer  = newContainer(flagDeployment)
	flagImage      string
	flagMount      = new(Path)
	flagInput      = new(Path)
	flagArgs       = newStringList(" ")
	flagCleanup    bool
	flagRollback   bool
	flagContinue   bool
	flagListen     = newHostPort("127.0.0.1", 0)
	flagVscode     bool
)

func patch(l *logrus.Entry, ns *Namespace, d *Deployment, p *Pod, c *Container, image, tag,
	mountPath string, rollback bool) (string, error) {
	if image == "" {
		image = c.getImage()
		tag = c.getTag()
	}

	if rollback {
		return configurd.Rollback(l, d.Resource())
	}
	return configurd.Patch(l, d.Resource(), c.Value(), image, tag, mountPath)
}

func shell(l *logrus.Entry, d *Deployment, p *Pod) (string, error) {
	return configurd.Shell(l, d.Resource(), p.Value())
}

func delve(l *logrus.Entry, d *Deployment, p *Pod, c *Container, input string, args []string,
	host string, port int, cleanup, dlvContinue, vscode bool) (string, error) {
	if cleanup {
		return configurd.DelveCleanup(l, d.Resource(), p.Value(), c.Value())
	}
	return configurd.Delve(l, d.Resource(), p.Value(), c.Value(), input, args, dlvContinue, host, port, vscode)
}
