package squadron

import (
	"bytes"
	b64 "encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/miracl/conflate"
	"github.com/pkg/errors"
)

func init() {
	// define the unmarshallers for the given file extensions, blank extension is the global unmarshaller
	conflate.Unmarshallers = conflate.UnmarshallerMap{
		".yaml": {conflate.YAMLUnmarshal},
		".yml":  {conflate.YAMLUnmarshal},
	}
}

type TemplateVars map[string]interface{}

func (tv *TemplateVars) add(name string, value interface{}) {
	(*tv)[name] = value
}

func executeFileTemplate(text string, templateVars interface{}, errorOnMissing bool) ([]byte, error) {
	templateFunctions := template.FuncMap{}
	templateFunctions["env"] = env
	templateFunctions["op"] = onePassword(templateVars, errorOnMissing)
	templateFunctions["base64"] = base64
	templateFunctions["default"] = defaultIndex
	templateFunctions["indent"] = indent
	templateFunctions["file"] = file(templateVars, errorOnMissing)
	templateFunctions["git"] = git

	tpl, err := template.New("squadron").Delims("<%", "%>").Funcs(templateFunctions).Parse(text)
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

func env(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", fmt.Errorf("env variable %q was empty", name)
	}
	return value, nil
}

func file(templateVars interface{}, errorOnMissing bool) func(v string) (string, error) {
	return func(v string) (string, error) {
		if v == "" {
			return "", nil
		} else if fileBytes, err := ioutil.ReadFile(v); err != nil {
			return "", errors.Wrap(err, "failed to read file")
		} else if renderedBytes, err := executeFileTemplate(string(fileBytes), templateVars, errorOnMissing); err != nil {
			return "", errors.Wrap(err, "failed to render file")
		} else {
			return string(bytes.TrimSpace(renderedBytes)), nil
		}
	}
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

func replace(in interface{}) {
	if value, ok := in.(map[string]interface{}); ok {
		for k, v := range value {
			if strings.Contains(k, "-") {
				value[strings.ReplaceAll(k, "-", "_")] = v
				delete(value, k)
			}
			replace(v)
		}
	}
}

func git(action string) (string, error) {
	cmd := exec.Command("git")

	switch action {
	case "tag":
		cmd.Args = append(cmd.Args, "describe", "--tags", "--always")
	case "commitsha":
		cmd.Args = append(cmd.Args, "rev-list", "-1", "HEAD")
	case "abbrevcommitsha":
		cmd.Args = append(cmd.Args, "rev-list", "-1", "HEAD", "--abbrev-commit")
	}
	res, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(bytes.TrimSpace(res)), nil
}

func indent(spaces int, v string) string {
	pad := strings.Repeat(" ", spaces)
	return strings.ReplaceAll(v, "\n", "\n"+pad)
}

func render(name, text string, data interface{}, errorOnMissing bool) (string, error) {
	var opts []string
	if !errorOnMissing {
		opts = append(opts, "missingkey=error")
	}
	out := bytes.NewBuffer([]byte{})
	if uuidTpl, err := template.New(name).Option(opts...).Parse(text); err != nil {
		return "", err
	} else if err := uuidTpl.Execute(out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}

func onePassword(templateVars interface{}, errorOnMissing bool) func(account, uuid, field string) (string, error) {
	return func(account, uuid, field string) (string, error) {
		if value, err := render("op", uuid, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			uuid = value
		}
		if value, err := render("op", field, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			field = value
		}

		// validate command
		if _, err := exec.LookPath("op"); err != nil {
			fmt.Println("Your templates includes a call to 1Password, please install it:")
			fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
			return "", err
		}

		// validate session
		if os.Getenv(fmt.Sprintf("OP_SESSION_%s", account)) == "" {
			if err := onePasswordSignIn(account); err != nil {
				return "", err
			}
		}

		res, err := onePasswordGet(uuid, field)
		if err != nil && strings.Contains(res, "You are not currently signed in") {
			// retry with login
			if err := onePasswordSignIn(account); err != nil {
				return "", err
			} else if res, err = onePasswordGet(uuid, field); err != nil {
				return "", err
			}
		} else if err != nil {
			return "", err
		}
		return res, nil
	}
}

func onePasswordGet(uuid, field string) (string, error) {
	res, err := exec.Command("op", "get", "item", uuid, "--fields", field).CombinedOutput()
	return string(res), err
}

func onePasswordSignIn(account string) error {
	fmt.Println("Your templates includes a call to 1Password, please sign to retrieve your session token:")

	// create command
	cmd := exec.Command("op", "signin", account, "--raw")

	// use multi writer to handle password prompt
	var stdoutBuf bytes.Buffer
	cmd.Stdout = io.MultiWriter(os.Stdout, &stdoutBuf)
	cmd.Stdin = os.Stdin

	// start the process and wait till it's finished
	if err := cmd.Start(); err != nil {
		return err
	} else if err := cmd.Wait(); err != nil {
		return err
	}

	if token := strings.TrimSuffix(stdoutBuf.String(), "\n"); token == "" {
		fmt.Printf("Failed to login into your '%s' account! Please refer to the manual:\n", account)
		fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
		return errors.New("failed to retrieve 1password session token")
	} else if err := os.Setenv(fmt.Sprintf("OP_SESSION_%s", account), token); err != nil {
		return err
	} else {
		fmt.Println("NOTE: If you want to skip this step, run:")
		fmt.Printf("export OP_SESSION_%s=%s\n", account, token)
	}

	return nil
}
