package config

import (
	"slices"
	"strings"
)

type Tags []Tag

func (t Tags) String() string {
	return strings.Join(t.Strings(), ",")
}

func (t Tags) SortedString() string {
	return strings.Join(t.SortedStrings(), ",")
}

func (t Tags) Strings() []string {
	ret := make([]string, len(t))
	for i, tag := range t {
		ret[i] = tag.String()
	}
	return ret
}

func (t Tags) SortedStrings() []string {
	ret := t.Strings()
	slices.Sort(ret)
	return ret
}
