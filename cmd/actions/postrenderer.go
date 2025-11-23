package actions

import (
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func NewPostRenderer(c *viper.Viper) *cobra.Command {
	cmd := &cobra.Command{
		Use:    "post-renderer [PATH]",
		Hidden: true,
		Args:   cobra.MinimumNArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			// this does the trick
			r, err := io.ReadAll(cmd.InOrStdin())
			if err != nil {
				return err
			}

			err = os.WriteFile(path.Join(args[0], ".chart.yaml"), r, 0600)
			if err != nil {
				return err
			}

			c := exec.CommandContext(cmd.Context(), "kustomize", "build", args[0]) //nolint:gosec
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr

			return c.Run()
		},
	}

	return cmd
}
