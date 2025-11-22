package git_test

import (
	"testing"

	"github.com/foomo/squadron/internal/git"
	"github.com/stretchr/testify/assert"
)

func TestOriginURL(t *testing.T) {
	urls := []string{
		"git@github.com:user/repo.git",
		"ssh://git@github.com/user/repo.git",
		"https://github.com/user/repo.git",
	}
	for _, url := range urls {
		assert.Equal(t, "https://github.com/user/repo", git.OriginURL(url))
	}
}
