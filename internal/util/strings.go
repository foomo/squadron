package util

import (
	"strings"
)

func StringToMap(s string) map[string]string {
	result := make(map[string]string)

	for line := range strings.SplitSeq(s, ",") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		k, v, _ := strings.Cut(line, "=")
		result[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	return result
}
