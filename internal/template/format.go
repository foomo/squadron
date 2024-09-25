package template

import (
	"fmt"
	"strings"
)

func quote(str ...any) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("%v", s))
		}
	}
	return "'" + strings.Join(out, " ") + "'"
}

func quoteAll(str ...any) string {
	out := make([]string, 0, len(str))
	for _, s := range str {
		if s != nil {
			out = append(out, fmt.Sprintf("'%v'", s))
		}
	}
	return strings.Join(out, " ")
}

func indent(spaces int, v string) string {
	pad := strings.Repeat("  ", spaces)
	return strings.ReplaceAll(v, "\n", "\n"+pad)
}
