package config

import (
	"slices"
	"strings"
)

// TagFilter is a parsed tag selector as provided through the `--tags` flag.
//
// Entries prefixed with `-` exclude a unit, all others include it. Includes are
// combined with OR, excludes take precedence over includes.
type TagFilter struct {
	include Tags
	exclude Tags
}

// NewTagFilter parses the given selector values into a TagFilter
func NewTagFilter(values []string) TagFilter {
	var ret TagFilter

	for _, value := range values {
		value = strings.TrimSpace(value)
		if after, ok := strings.CutPrefix(value, "-"); ok {
			if after != "" && !slices.Contains(ret.exclude, Tag(after)) {
				ret.exclude = append(ret.exclude, Tag(after))
			}
		} else if value != "" && !slices.Contains(ret.include, Tag(value)) {
			ret.include = append(ret.include, Tag(value))
		}
	}

	return ret
}

// Empty returns true if the filter does not select anything
func (f TagFilter) Empty() bool {
	return len(f.include) == 0 && len(f.exclude) == 0
}

// Referenced returns all tags referenced by the filter
func (f TagFilter) Referenced() Tags {
	ret := make(Tags, 0, len(f.include)+len(f.exclude))
	ret = append(ret, f.include...)

	for _, tag := range f.exclude {
		if !slices.Contains(ret, tag) {
			ret = append(ret, tag)
		}
	}

	return ret
}

// Match returns true if the given tags satisfy the filter
func (f TagFilter) Match(tags Tags) bool {
	for _, tag := range f.exclude {
		if slices.Contains(tags, tag) {
			return false
		}
	}

	if len(f.include) == 0 {
		return true
	}

	for _, tag := range f.include {
		if slices.Contains(tags, tag) {
			return true
		}
	}

	return false
}
