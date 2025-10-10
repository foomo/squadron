package util

import (
	"bytes"
	"text/template"

	"github.com/pkg/errors"
)

func RenderTemplateString(s string, data any) (string, error) {
	t, err := template.New("template").Parse(s)
	if err != nil {
		return "", errors.Wrap(err, "failed to parse template")
	}

	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return "", errors.Wrap(err, "failed to execute template")
	}

	return out.String(), nil
}
