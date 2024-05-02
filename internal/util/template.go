package util

import (
	"bytes"
	"text/template"
)

func RenderTemplateString(s string, data any) (string, error) {
	t, err := template.New("template").Parse(s)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}
