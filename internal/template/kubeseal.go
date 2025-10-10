package template

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	"github.com/pterm/pterm"
)

func kubeseal(ctx context.Context) func(values ...string) (string, error) {
	return func(values ...string) (string, error) {
		var value string

		if len(values) == 0 {
			return "", errors.Errorf("missing value")
		} else if len(values) == 1 {
			value = values[0]
		} else {
			value, values = values[len(values)-1], values[:len(values)-1]
		}

		cmd := exec.CommandContext(ctx, "kubeseal", "--raw", "--from-file=/dev/stdin")
		cmd.Args = append(cmd.Args, values...)
		cmd.Env = os.Environ()
		cmd.Stdin = bytes.NewReader([]byte(value))

		if v := os.Getenv("SQUADRON_KUBESEAL_NAME"); v != "" {
			cmd.Args = append(cmd.Args, "--name", v)
		}

		if v := os.Getenv("SQUADRON_KUBESEAL_NAMESPACE"); v != "" {
			cmd.Args = append(cmd.Args, "--namespace", v)
		}

		if v := os.Getenv("SQUADRON_KUBESEAL_CONTROLLER_NAME"); v != "" {
			cmd.Args = append(cmd.Args, "--controller-name", v)
		}

		if v := os.Getenv("SQUADRON_KUBESEAL_CONTROLLER_NAMESPACE"); v != "" {
			cmd.Args = append(cmd.Args, "--controller-namespace", v)
		}

		if v := os.Getenv("SQUADRON_KUBESEAL_EXTRA_ARGS"); v != "" {
			cmd.Args = append(cmd.Args, strings.Split(v, " ")...)
		}

		res, err := cmd.CombinedOutput()
		if err != nil {
			pterm.Debug.Println(cmd.String())
			pterm.Error.Println(string(res))

			return "", err
		}

		return string(bytes.Trim(bytes.TrimSpace(res), "\n")), nil
	}
}
