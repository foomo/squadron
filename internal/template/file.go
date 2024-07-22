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
		} else if fileBytes, err := os.ReadFile(v); err != nil {
			return "", errors.Wrap(err, "failed to read file")
		} else if renderedBytes, err := ExecuteFileTemplate(ctx, string(fileBytes), templateVars, errorOnMissing); err != nil {
			return "", errors.Wrap(err, "failed to render file")
		} else {
			return string(bytes.TrimSpace(renderedBytes)), nil
		}
	}
}
