package util

import (
	"strings"
)

func ToSnakeCaseKeys(in interface{}) {
	if value, ok := in.(map[string]interface{}); ok {
		for k, v := range value {
			if strings.Contains(k, "-") {
				value[strings.ReplaceAll(k, "-", "_")] = v
				delete(value, k)
			}
			ToSnakeCaseKeys(v)
		}
	}
}
