package template

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"text/template"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"
)

var (
	onePasswordCache map[string]map[string]string
	onePasswordUUID  = regexp.MustCompile(`^[a-z0-9]{26}$`)
)

var ErrOnePasswordNotSignedIn = errors.New("not signed in")

func onePasswordConnectGet(client connect.Client, vaultUUID, itemUUID string) (map[string]string, error) {
	var item *onepassword.Item
	if onePasswordUUID.MatchString(itemUUID) {
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

func onePasswordConnectGetDocument(client connect.Client, vaultUUID, itemUUID string) (string, error) {
	var item *onepassword.Item
	if onePasswordUUID.MatchString(itemUUID) {
		if v, err := client.GetItem(itemUUID, vaultUUID); err != nil {
			return "", err
		} else {
			item = v
		}
	} else {
		if v, err := client.GetItemByTitle(itemUUID, vaultUUID); err != nil {
			return "", err
		} else {
			item = v
		}
	}

	if item.Category != onepassword.Document {
		return "", errors.Errorf("unexpected document type: %s", item.Category)
	} else if len(item.Files) != 0 {
		return "", errors.Errorf("unexpected document files length: %d", len(item.Files))
	}

	res, err := client.GetFileContent(item.Files[0])
	if err != nil {
		return "", err
	}

	return strings.Trim(string(res), "\n"), nil
}

var onePasswordGetLock sync.Mutex

func onePasswordGet(ctx context.Context, account, vaultUUID, itemUUID string) (map[string]string, error) {
	onePasswordGetLock.Lock()
	defer onePasswordGetLock.Unlock()
	var v struct {
		Vault struct {
			ID string `json:"id"`
		} `json:"vault"`
		Fields []struct {
			ID    string `json:"id"`
			Type  string `json:"type"` // CONCEALED, STRING
			Label string `json:"label"`
			Value any    `json:"value"`
		} `json:"fields"`
	}
	if res, err := exec.CommandContext(ctx, "op", "item", "get", itemUUID, "--vault", vaultUUID, "--account", account, "--format", "json").CombinedOutput(); err != nil && strings.Contains(string(res), "You are not currently signed in") {
		return nil, ErrOnePasswordNotSignedIn
	} else if err != nil {
		return nil, errors.Wrap(err, string(res))
	} else if err := json.Unmarshal(res, &v); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal secret")
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

var onePasswordGetDocumentLock sync.Mutex

func onePasswordGetDocument(ctx context.Context, account, vaultUUID, itemUUID string) (string, error) {
	onePasswordGetDocumentLock.Lock()
	defer onePasswordGetDocumentLock.Unlock()
	res, err := exec.CommandContext(ctx, "op", "document", "get", itemUUID, "--vault", vaultUUID, "--account", account).CombinedOutput()
	if err != nil && strings.Contains(string(res), "You are not currently signed in") {
		return "", ErrOnePasswordNotSignedIn
	} else if err != nil {
		return "", errors.Wrap(err, string(res))
	}
	return strings.Trim(string(res), "\n"), nil
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

func isConnect() bool {
	return os.Getenv("OP_CONNECT_HOST") != "" && os.Getenv("OP_CONNECT_TOKEN") != ""
}

func isServiceAccount() bool {
	return os.Getenv("OP_SERVICE_ACCOUNT_TOKEN") != ""
}

var onePasswordInitLock sync.Mutex

func onePasswordInit(ctx context.Context, account string) error {
	onePasswordInitLock.Lock()
	defer onePasswordInitLock.Unlock()

	// validate cache
	if onePasswordCache != nil {
		return nil
	}

	onePasswordCache = map[string]map[string]string{}

	// validate env
	if isConnect() || isServiceAccount() {
		return nil
	}

	// validate executeable
	if _, err := exec.LookPath("op"); err != nil {
		pterm.Warning.Println("Your templates includes a call to 1Password, please install it:")
		pterm.Warning.Println("https://support.1password.com/command-line-getting-started/#set-up-the-command-line-tool")
		return errors.Wrap(err, "failed to lookup op")
	}

	// validate auth
	if _, err := exec.CommandContext(ctx, "op", "account", "get", "--account", account).CombinedOutput(); err == nil {
		return nil
	}

	// validate auth env
	if os.Getenv(fmt.Sprintf("OP_SESSION_%s", account)) == "" {
		if err := onePasswordSignIn(ctx, account); err != nil {
			return errors.Wrap(err, "failed to sign in")
		}
	}

	return nil
}

func onePassword(ctx context.Context, templateVars any, errorOnMissing bool) func(account, vaultUUID, itemUUID, field string) (string, error) {
	return func(account, vaultUUID, itemUUID, field string) (string, error) {
		// init
		if err := onePasswordInit(ctx, account); err != nil {
			return "", err
		}
		// render uuid & field params
		if value, err := onePasswordRender("op", itemUUID, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			itemUUID = value
		}
		if value, err := onePasswordRender("op", field, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			field = value
		}

		// create cache key
		cacheKey := strings.Join([]string{account, vaultUUID, itemUUID}, "#")

		if _, ok := onePasswordCache[cacheKey]; !ok {
			if isConnect() {
				client, err := connect.NewClientFromEnvironment()
				if err != nil {
					return "", err
				}
				if res, err := onePasswordConnectGet(client, vaultUUID, itemUUID); err != nil {
					return "", err
				} else {
					onePasswordCache[cacheKey] = res
				}
			} else {
				if res, err := onePasswordGet(ctx, account, vaultUUID, itemUUID); err != nil {
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

func onePasswordDocument(ctx context.Context, templateVars any, errorOnMissing bool) func(account, vaultUUID, itemUUID string) (string, error) {
	return func(account, vaultUUID, itemUUID string) (string, error) {
		// init
		if err := onePasswordInit(ctx, account); err != nil {
			return "", err
		}

		// render uuid & field params
		if value, err := onePasswordRender("op", itemUUID, templateVars, errorOnMissing); err != nil {
			return "", err
		} else {
			itemUUID = value
		}

		// create cache key
		cacheKey := strings.Join([]string{account, vaultUUID, itemUUID}, "#")

		if _, ok := onePasswordCache[cacheKey]; !ok {
			if isConnect() {
				if client, err := connect.NewClientFromEnvironment(); err != nil {
					return "", err
				} else if res, err := onePasswordConnectGetDocument(client, vaultUUID, itemUUID); err != nil {
					return "", err
				} else {
					onePasswordCache[cacheKey] = map[string]string{"document": res}
				}
			} else {
				if res, err := onePasswordGetDocument(ctx, account, vaultUUID, itemUUID); err != nil {
					return "", err
				} else {
					onePasswordCache[cacheKey] = map[string]string{"document": res}
				}
			}
		}

		if value, ok := onePasswordCache[cacheKey]["document"]; !ok {
			return "", nil
		} else {
			return value, nil
		}
	}
}

func onePasswordRender(name, text string, data any, errorOnMissing bool) (string, error) {
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
