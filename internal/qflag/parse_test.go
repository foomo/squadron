package qflag_test

import (
	"testing"
	"time"

	"github.com/foomo/squadron/internal/qflag"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		f := pflag.NewFlagSet("build", pflag.ContinueOnError)
		f.Int("int", 0, "")
		f.Bool("bool", false, "")
		f.Float32("float", 0, "")
		f.String("string", "", "")
		f.Duration("duration", time.Duration(0), "")
		f.StringArray("string-array", nil, "")
		f.StringSlice("string-slice", nil, "")

		args := qflag.Parse(f)
		assert.Empty(t, args)
	})

	t.Run("defaults", func(t *testing.T) {
		f := pflag.NewFlagSet("build", pflag.ContinueOnError)
		f.Int("int", 1, "")
		f.Float32("float", 1, "")
		f.Bool("bool", true, "")
		f.String("string", "foo", "")
		f.Duration("duration", time.Second, "")
		f.StringArray("string-array", []string{"foo"}, "")
		f.StringSlice("string-slice", []string{"foo"}, "")

		args := qflag.Parse(f)
		assert.Len(t, args, 13)
		t.Log(args)
	})

	t.Run("overrides", func(t *testing.T) {
		f := pflag.NewFlagSet("build", pflag.ContinueOnError)
		f.Int("int", 1, "")
		f.Float32("float", 1, "")
		f.Bool("bool", false, "")
		f.String("string", "foo", "")
		f.Duration("duration", time.Second, "")
		f.StringArray("string-array", []string{"foo"}, "")
		f.StringSlice("string-slice", []string{"foo"}, "")

		err := f.Parse([]string{
			"--bool", "2",
			"--int", "2",
			"--float", "2",
			"--string", "baz",
			"--duration", "2s",
			"--string-array", "baz",
			"--string-slice", "baz",
		})
		require.NoError(t, err)

		args := qflag.Parse(f)
		assert.Len(t, args, 13)
		t.Log(args)
	})
}
