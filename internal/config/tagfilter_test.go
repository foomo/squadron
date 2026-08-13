package config_test

import (
	"testing"

	testingx "github.com/foomo/go/testing"
	tagx "github.com/foomo/go/testing/tag"
	"github.com/foomo/squadron/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewTagFilter(t *testing.T) {
	testingx.Tags(t, tagx.Short)
	t.Parallel()

	tests := []struct {
		name       string
		values     []string
		empty      bool
		referenced config.Tags
	}{
		{
			name:       "nil",
			values:     nil,
			empty:      true,
			referenced: config.Tags{},
		},
		{
			name:       "blank values",
			values:     []string{"", " ", "-"},
			empty:      true,
			referenced: config.Tags{},
		},
		{
			name:       "include",
			values:     []string{"frontend"},
			referenced: config.Tags{"frontend"},
		},
		{
			name:       "whitespace",
			values:     []string{"frontend", " backend", "-skip "},
			referenced: config.Tags{"frontend", "backend", "skip"},
		},
		{
			name:       "duplicates",
			values:     []string{"frontend", "frontend", "-skip", "-skip"},
			referenced: config.Tags{"frontend", "skip"},
		},
		{
			name:       "include and exclude of the same tag",
			values:     []string{"skip", "-skip"},
			referenced: config.Tags{"skip"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			filter := config.NewTagFilter(test.values)
			assert.Equal(tt, test.empty, filter.Empty())
			assert.Equal(tt, test.referenced, filter.Referenced())
		})
	}
}

func TestTagFilter_Match(t *testing.T) {
	testingx.Tags(t, tagx.Short)
	t.Parallel()

	var (
		frontend = config.Tags{"frontend"}
		backend  = config.Tags{"backend", "skip"}
		untagged = config.Tags{}
	)

	tests := []struct {
		name   string
		values []string
		want   map[string]bool
	}{
		{
			name:   "empty filter matches everything",
			values: nil,
			want:   map[string]bool{"frontend": true, "backend": true, "untagged": true},
		},
		{
			name:   "single include",
			values: []string{"frontend"},
			want:   map[string]bool{"frontend": true, "backend": false, "untagged": false},
		},
		{
			name:   "multiple includes are combined with or",
			values: []string{"frontend", "backend"},
			want:   map[string]bool{"frontend": true, "backend": true, "untagged": false},
		},
		{
			name:   "exclude only",
			values: []string{"-skip"},
			want:   map[string]bool{"frontend": true, "backend": false, "untagged": true},
		},
		{
			name:   "exclude takes precedence over include",
			values: []string{"backend", "-skip"},
			want:   map[string]bool{"frontend": false, "backend": false, "untagged": false},
		},
		{
			name:   "include and exclude combined",
			values: []string{"frontend", "backend", "-skip"},
			want:   map[string]bool{"frontend": true, "backend": false, "untagged": false},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			tt.Parallel()

			filter := config.NewTagFilter(test.values)
			assert.Equal(tt, test.want["frontend"], filter.Match(frontend), "frontend")
			assert.Equal(tt, test.want["backend"], filter.Match(backend), "backend")
			assert.Equal(tt, test.want["untagged"], filter.Match(untagged), "untagged")
		})
	}
}
