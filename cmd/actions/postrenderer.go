package actions

import (
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
)

func init() {
}

var postRendererCmd = &cobra.Command{
	Use:     "post-renderer [PATH]",
	Hidden:  true,
	Short:   "render chart templates locally and display the output",
	Example: "  squadron template storefinder frontend backend --namespace demo",
	Args:    cobra.MinimumNArgs(0),
	RunE: func(cmd *cobra.Command, args []string) error {

		// this does the trick
		r, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return err
		}

		err = os.WriteFile(path.Join(args[0], ".chart.yaml"), r, 0644)
		if err != nil {
			return err
		}

		c := exec.CommandContext(cmd.Context(), "kustomize", "build", args[0])
		c.Stdout = os.Stdout
		return c.Run()
	},
}
