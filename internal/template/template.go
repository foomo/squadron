package template

import (
	"bytes"
	"context"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

func ExecuteFileTemplate(ctx context.Context, text string, templateVars any, errorOnMissing bool) ([]byte, error) {
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	funcMap["env"] = env
	funcMap["quote"] = quote
	funcMap["quoteAll"] = quoteAll
	funcMap["envDefault"] = envDefault

	// deprecated
	funcMap["indent"] = indent
	funcMap["base64"] = base64
	funcMap["defaultIndex"] = defaultIndexValue

	funcMap["op"] = onePassword(ctx, templateVars, errorOnMissing)
	funcMap["git"] = git(ctx)
	funcMap["opDoc"] = onePasswordDocument(ctx, templateVars, errorOnMissing)
	funcMap["file"] = file(ctx, templateVars, errorOnMissing)
	funcMap["kubeseal"] = kubeseal(ctx)

	funcMap["toToml"] = toTOML
	funcMap["fromToml"] = fromTOML
	funcMap["toYaml"] = toYAML
	funcMap["toYamlPretty"] = toYAMLPretty
	funcMap["fromYaml"] = fromYAML
	funcMap["fromYamlArray"] = fromYAMLArray
	funcMap["toJson"] = toJSON
	funcMap["fromJson"] = fromJSON
	funcMap["fromJsonArray"] = fromJSONArray

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
