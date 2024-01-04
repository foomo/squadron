package template

import (
	"strings"
)

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
