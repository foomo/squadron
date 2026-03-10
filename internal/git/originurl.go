package git

import (
	"strings"
)

// OriginURL converts SSH git URLs to HTTPS format.
func OriginURL(origin string) string {
	if strings.HasPrefix(origin, "git@") {
		// Format is git@github.com:user/repo.git
		parts := strings.SplitN(origin, ":", 2)
		if len(parts) == 2 {
			origin = "https://" + strings.TrimPrefix(parts[0], "git@") + "/" + parts[1]
		}
	}

	if after, ok := strings.CutPrefix(origin, "ssh://git@"); ok {
		// Format is ssh://git@github.com/user/repo.git
		origin = "https://" + after
	}

	// Not an SSH URL, return unchanged
	return strings.TrimSuffix(origin, ".git")
}
