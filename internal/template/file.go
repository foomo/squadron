package template

import (
	"bytes"
	"context"
	"os"

	"github.com/pkg/errors"
)

func file(ctx context.Context, templateVars any, errorOnMissing bool) func(v string) (string, error) {
	return func(v string) (string, error) {
		if v == "" {
			return "", nil
		}

		fileBytes, err := os.ReadFile(os.ExpandEnv(v))
		if err != nil {
			return "", errors.Wrap(err, "failed to read file")
		}

		renderedBytes, err := ExecuteFileTemplate(ctx, string(fileBytes), templateVars, errorOnMissing)
		if err != nil {
			return "", errors.Wrap(err, "failed to render file")
		}

		return string(bytes.TrimSpace(renderedBytes)), nil
	}
}
