package template

import (
	b64 "encoding/base64"
)

func base64(v string) string {
	return b64.StdEncoding.EncodeToString([]byte(v))
}
