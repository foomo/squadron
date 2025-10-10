package template

import (
	"bytes"
	"context"
	"os/exec"
)

func git(ctx context.Context) func(action string) (string, error) {
	return func(action string) (string, error) {
		cmd := exec.CommandContext(ctx, "git")

		switch action {
		case "commitsha":
			cmd.Args = append(cmd.Args, "rev-list", "-1", "HEAD")
		case "abbrevcommitsha":
			cmd.Args = append(cmd.Args, "rev-list", "-1", "HEAD", "--abbrev-commit")
		default:
			cmd.Args = append(cmd.Args, "describe", "--tags", "--always")
		}

		res, err := cmd.Output()
		if err != nil {
			return "", err
		}

		return string(bytes.TrimSpace(res)), nil
	}
}
