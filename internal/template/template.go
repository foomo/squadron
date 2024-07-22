package template

import (
	"bytes"
	"context"
	"text/template"
)

func ExecuteFileTemplate(ctx context.Context, text string, templateVars any, errorOnMissing bool) ([]byte, error) {
	funcMap := template.FuncMap{
		"env":          env,
		"envDefault":   envDefault,
		"op":           onePassword(ctx, templateVars, errorOnMissing),
		"opDoc":        onePasswordDocument(ctx, templateVars, errorOnMissing),
		"base64":       base64,
		"default":      defaultValue,
		"defaultIndex": defaultIndexValue,
		"indent":       indent,
		"replace":      replace,
		"file":         file(ctx, templateVars, errorOnMissing),
		"git":          git(ctx),
		"quote":        quote,
	}
	tpl, err := template.New("squadron").Delims("<% ", " %>").Funcs(funcMap).Parse(text)
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer([]byte{})
	if errorOnMissing {
		tpl = tpl.Option("missingkey=error")
	}
	if err := tpl.Funcs(funcMap).Execute(out, templateVars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
