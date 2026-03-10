package util

import (
	"strings"
)

func StringToMap(key string) map[string]string {
	result := make(map[string]string)

	for _, line := range strings.Split(key, ",") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		k, v, _ := strings.Cut(line, "=")
		result[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	return result
}
