package template

import (
	"strings"
)

func indent(spaces int, v string) string {
	pad := strings.Repeat("  ", spaces)
	return strings.ReplaceAll(v, "\n", "\n"+pad)
}
