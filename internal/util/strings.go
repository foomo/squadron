package util

import (
	"strings"
)

func ToSnakeCaseKeys(in any) {
	if value, ok := in.(map[string]any); ok {
		for k, v := range value {
			if strings.Contains(k, "-") {
				value[strings.ReplaceAll(k, "-", "_")] = v
				delete(value, k)
			}
			ToSnakeCaseKeys(v)
		}
	}
}
