package actions

import (
	"github.com/foomo/configurd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build [SERVICE] -t {TAG}",
	Short: "Build a service with a given tag",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := build(args[0], flagTag, flagDir, flagVerbose)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func build(service, tag, dir string, verbose bool) (string, error) {
	cnf := mustNewConfigurd(configurd.Config{
		Log:      logrus.New(),
		Tag:      tag,
		BasePath: dir,
		Verbose:  verbose,
	})

	return cnf.Build(service)
}
