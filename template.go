package squadron

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
	templateFunctions["op"] = onePassword
	templateFunctions["base64"] = base64
	templateFunctions["default"] = defaultIndex
	templateFunctions["yaml"] = yamlFile
	templateFunctions["indent"] = indent

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

func yamlFile(v string) (string, error) {
	var bs []byte
	var err error
	bs, err = ioutil.ReadFile(v)
	if err != nil {
		return fmt.Sprintf("%q", v), err
	}
	return strings.Trim(string(bs), "\n"), nil
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

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return strings.Replace(v, "\n", "\n"+pad, -1)
}

func onePassword(account, uuid, field string) (string, error) {
	// validate command
	if _, err := exec.LookPath("op"); err != nil {
		fmt.Println("Your templates includes a call to 1Password, please install it:")
		fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
		return "", err
	}

	// validate session
	if os.Getenv(fmt.Sprintf("OP_SESSION_%s", account)) == "" {
		fmt.Println("Your templates includes a call to 1Password, please sign into your account:")
		if token, err := exec.Command("op", "signin", account, "--raw").Output(); err != nil {
			fmt.Println(fmt.Sprintf("Failed to login into your '%s' account! Please refer to the manual:", account))
			fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
			return "", err
		} else if err := os.Setenv(fmt.Sprintf("OP_SESSION_%s", account), string(token)); err != nil {
			return "", err
		} else {
			fmt.Println("NOTE: If you want to skip this step, run:")
			fmt.Println(fmt.Sprintf("eval $(op signin %s)", account))
		}
	}
	res, err := exec.Command("op", "get", "item", uuid, "--fields", field).Output()
	if err != nil {
		return "", err
	}
	return string(res), nil
}
