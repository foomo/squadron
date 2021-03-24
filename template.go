package squadron

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/foomo/config-bob/builder"
	"github.com/imdario/mergo"
	"gopkg.in/yaml.v3"
)

type TemplateVars map[string]interface{}

func (tv *TemplateVars) add(name string, value interface{}) {
	(*tv)[name] = value
}

func executeFileTemplate(path string, templateVars interface{}, errorOnMissing bool) ([]byte, error) {
	templateFunctions := template.FuncMap{}
	templateFunctions["env"] = builder.TemplateFuncs["env"]
	templateFunctions["op"] = builder.TemplateFuncs["op"]
	templateFunctions["base64"] = base64
	templateFunctions["default"] = defaultIndex
	templateFunctions["yaml"] = yamlMixed
	// todo test yaml

	templateBytes, errRead := ioutil.ReadFile(path)
	if errRead != nil {
		return nil, errRead
	}
	tpl, err := template.New("squadron").Funcs(templateFunctions).Parse(string(templateBytes))
	if err != nil {
		return nil, err
	}
	out := bytes.NewBuffer([]byte{})
	if errorOnMissing {
		tpl = tpl.Option("missingkey=error")
	}
	if err := tpl.Funcs(templateFunctions).Execute(out, templateVars); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func yamlMixed(v interface{}) (string, error) {
	if vString, ok := v.(string); ok {
		var err error
		v, err = os.ReadFile(vString)
		if err != nil {
			return fmt.Sprintf("%q", v), err
		}
	}
	yamlBytes, err := yaml.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%q", v), err
	}
	return strings.Trim(string(yamlBytes), "\n"), nil
}

func defaultIndex(v map[string]interface{}, index string, def interface{}) interface{} {
	var ok bool
	if _, ok = v[index]; ok {
		return v[index]
	}
	return def
}

func base64(v string) string {
	return b64.StdEncoding.EncodeToString([]byte(v))
}

func executeSquadronTemplate(file string, c *Configuration, tv TemplateVars) error {
	// execute without errors to get existing values
	out, err := executeFileTemplate(file, tv, false)
	if err != nil {
		return err
	}
	var vars map[string]interface{}
	if err := yaml.Unmarshal(out, &vars); err != nil {
		return err
	}
	// execute again with loaded template vars
	tv.add("Squadron", vars["squadron"])
	out, err = executeFileTemplate(file, tv, true)
	if err != nil {
		return err
	}
	if err := yaml.Unmarshal(out, &c); err != nil {
		return err
	}
	return nil
}

func mergeSquadronFiles(files []string, c *Configuration, tv TemplateVars) error {
	var mcs []Configuration
	for _, f := range files {
		mc := Configuration{}
		if err := executeSquadronTemplate(f, &mc, tv); err != nil {
			return err
		}
		mcs = append(mcs, mc)
	}
	for _, mc := range mcs {
		if err := mergo.Merge(c, mc, mergo.WithOverride); err != nil {
			return err
		}
	}
	return nil
}
