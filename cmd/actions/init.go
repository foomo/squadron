package actions

import (
	"path"

	"github.com/foomo/configurd"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [NAME]",
	Short: "Initialize an example application with configurd",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		_, err := Init(args[0], flagDir, flagVerbose)
		if err != nil {
			log.Fatal(err)
		}
	},
}

func Init(name, dir string, flagVerbose bool) (string, error) {
	output, err := configurd.Init(log, path.Join(dir, name), flagVerbose)
	if err != nil {
		return "", outputErrorf(output, err, "could not initialize an example configuraiton")
	}
	return output, nil
}
