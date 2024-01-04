package config_test

import (
	"testing"

	"github.com/foomo/squadron/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestMap_Trim(t *testing.T) {
	tests := []struct {
		name  string
		value config.Map[any]
		want  int
	}{
		{
			name: "string",
			value: config.Map[any]{
				"foo": "bar",
				"baz": "",
			},
			want: 1,
		},
		{
			name: "slice",
			value: config.Map[any]{
				"foo": []string{"foo"},
				"baz": []string{},
				"bar": nil,
			},
			want: 1,
		},
		{
			name: "map",
			value: config.Map[any]{
				"foo": map[string]string{"foo": "foo"},
				"baz": map[string]string{},
				"bar": nil,
			},
			want: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(tt *testing.T) {
			test.value.Trim()
			assert.Len(tt, test.value, test.want)
		})
	}
}
