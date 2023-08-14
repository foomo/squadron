package squadron

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/miracl/conflate"
	"github.com/pkg/errors"
)

func init() {
	// define the unmarshallers for the given file extensions, blank extension is the global unmarshaller
	conflate.Unmarshallers = conflate.UnmarshallerMap{
		".yaml": conflate.UnmarshallerFuncs{conflate.YAMLUnmarshal},
		".yml":  conflate.UnmarshallerFuncs{conflate.YAMLUnmarshal},
	}
}

type TemplateVars map[string]interface{}

func (tv *TemplateVars) add(name string, value interface{}) {
	(*tv)[name] = value
}

func executeFileTemplate(ctx context.Context, text string, templateVars interface{}, errorOnMissing bool) ([]byte, error) {
	templateFunctions := template.FuncMap{}
	templateFunctions["env"] = env
	templateFunctions["envDefault"] = envDefault
	templateFunctions["op"] = onePassword(ctx, templateVars, errorOnMissing)
	templateFunctions["base64"] = base64
	templateFunctions["default"] = defaultValue
	templateFunctions["defaultIndex"] = defaultIndexValue
	templateFunctions["indent"] = indent
	templateFunctions["replace"] = replace
	templateFunctions["file"] = file(ctx, templateVars, errorOnMissing)
	templateFunctions["git"] = git(ctx)
	templateFunctions["quote"] = quote

	tpl, err := template.New("squadron").Delims("<% ", " %>").Funcs(templateFunctions).Parse(text)
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
	if value := os.Getenv(name); value == "" {
		return "", fmt.Errorf("env variable %q was empty", name)
	} else {
		return value, nil
	}
}

func envDefault(name, fallback string) (string, error) {
	if value := os.Getenv(name); value == "" {
		return fallback, nil
	} else {
		return value, nil
	}
}

func file(ctx context.Context, templateVars interface{}, errorOnMissing bool) func(v string) (string, error) {
	return func(v string) (string, error) {
		if v == "" {
			return "", nil
		} else if fileBytes, err := os.ReadFile(v); err != nil {
			return "", errors.Wrap(err, "failed to read file")
		} else if renderedBytes, err := executeFileTemplate(ctx, string(fileBytes), templateVars, errorOnMissing); err != nil {
			return "", errors.Wrap(err, "failed to render file")
		} else {
			return string(bytes.TrimSpace(renderedBytes)), nil
		}
	}
}

func defaultValue(value string, def interface{}) interface{} {
	if value == "" {
		return def
	}
	return value
}

func defaultIndexValue(v map[string]interface{}, index string, def interface{}) interface{} {
	var ok bool
	if _, ok = v[index]; ok {
		return v[index]
	}
	return def
}

func base64(v string) string {
	return b64.StdEncoding.EncodeToString([]byte(v))
}

func toSnakeCaseKeys(in interface{}) {
	if value, ok := in.(map[string]interface{}); ok {
		for k, v := range value {
			if strings.Contains(k, "-") {
				value[strings.ReplaceAll(k, "-", "_")] = v
				delete(value, k)
			}
			toSnakeCaseKeys(v)
		}
	}
}

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

func indent(spaces int, v string) string {
	pad := strings.Repeat("  ", spaces)
	return strings.ReplaceAll(v, "\n", "\n"+pad)
}

func quote(v string) string {
	return "'" + v + "'"
}

func replace(old, new, v string) string {
	return strings.ReplaceAll(v, old, new)
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

var (
	onePasswordCache map[string]map[string]string
	onePasswordUUID  = regexp.MustCompile(`^[a-z0-9]{26}$`)
)

func onePassword(ctx context.Context, templateVars interface{}, errorOnMissing bool) func(account, vaultUUID, itemUUID, field string) (string, error) {
	if onePasswordCache == nil {
		onePasswordCache = map[string]map[string]string{}
	}
	return func(account, vaultUUID, itemUUID, field string) (string, error) {
		// validate command
		if mode := os.Getenv("OP_MODE"); mode == "ci" {
			// do nothing
		} else if _, err := exec.LookPath("op"); err != nil {
			fmt.Println("Your templates includes a call to 1Password, please install it:")
			fmt.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
			return "", err
		} else if os.Getenv(fmt.Sprintf("OP_SESSION_%s", account)) == "" {
			if err := onePasswordSignIn(ctx, account); err != nil {
				return "", err
			}
		}

		// render uuid & field params
		if value, err := render("op", itemUUID, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			itemUUID = value
		}
		if value, err := render("op", field, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			field = value
		}

		// create cache key
		cacheKey := strings.Join([]string{account, vaultUUID, itemUUID}, "#")

		if mode := os.Getenv("OP_MODE"); mode == "ci" {
			if _, ok := onePasswordCache[cacheKey]; !ok {
				if client, err := connect.NewClientFromEnvironment(); err != nil {
					return "", err
				} else if res, err := onePasswordCIGet(client, vaultUUID, itemUUID); err != nil {
					return "", err
				} else {
					onePasswordCache[cacheKey] = res
				}
			}
		} else {
			if _, ok := onePasswordCache[cacheKey]; !ok {
				if res, err := onePasswordGet(ctx, vaultUUID, itemUUID); !errors.Is(err, ErrOnePasswordNotSignedIn) {
					if err != nil {
						return "", err
					} else {
						onePasswordCache[cacheKey] = res
					}
				} else if err := onePasswordSignIn(ctx, account); err != nil {
					return "", err
				} else if res, err = onePasswordGet(ctx, vaultUUID, itemUUID); err != nil {
					return "", err
				} else {
					onePasswordCache[cacheKey] = res
				}
			}
		}

		if value, ok := onePasswordCache[cacheKey][field]; !ok {
			return "", nil
		} else {
			return value, nil
		}
	}
}

var ErrOnePasswordNotSignedIn = errors.New("not signed in")

func onePasswordCIGet(client connect.Client, vaultUUID, itemUUID string) (map[string]string, error) {
	var item *onepassword.Item
	if onePasswordUUID.Match([]byte(itemUUID)) {
		if v, err := client.GetItem(itemUUID, vaultUUID); err != nil {
			return nil, err
		} else {
			item = v
		}
	} else {
		if v, err := client.GetItemByTitle(itemUUID, vaultUUID); err != nil {
			return nil, err
		} else {
			item = v
		}
	}

	ret := map[string]string{}
	for _, f := range item.Fields {
		ret[f.Label] = f.Value
	}

	return ret, nil
}

func onePasswordGet(ctx context.Context, vaultUUID string, itemUUID string) (map[string]string, error) {
	var v struct {
		Vault struct {
			ID string `json:"id"`
		} `json:"vault"`
		Fields []struct {
			ID    string      `json:"id"`
			Type  string      `json:"type"` // CONCEALED, STRING
			Label string      `json:"label"`
			Value interface{} `json:"value"`
		} `json:"fields"`
	}
	if res, err := exec.CommandContext(ctx, "op", "item", "get", itemUUID, "--format", "json").CombinedOutput(); err != nil && strings.Contains(string(res), "You are not currently signed in") {
		return nil, ErrOnePasswordNotSignedIn
	} else if err != nil {
		return nil, err
	} else if err := json.Unmarshal(res, &v); err != nil {
		return nil, err
	} else if v.Vault.ID != vaultUUID {
		return nil, errors.Errorf("wrong vault UUID %s for item %s", vaultUUID, itemUUID)
	} else {
		ret := map[string]string{}
		aliases := map[string]string{
			"notesPlain": "notes",
		}
		for _, field := range v.Fields {
			if alias, ok := aliases[field.Label]; ok {
				ret[alias] = fmt.Sprintf("%v", field.Value)
			} else {
				ret[field.Label] = fmt.Sprintf("%v", field.Value)
			}
		}
		return ret, nil
	}
}

func onePasswordSignIn(ctx context.Context, account string) error {
	fmt.Println("Your templates includes a call to 1Password, please sign in to retrieve your session token:")

	// create command
	cmd := exec.CommandContext(ctx, "op", "signin", account, "--raw")

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
