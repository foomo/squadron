package qflag

import (
	"slices"
	"strings"

	"github.com/spf13/pflag"
)

func Parse(fs *pflag.FlagSet) []string {
	var ret []string

	fs.VisitAll(func(f *pflag.Flag) {
		switch {
		case slices.Contains([]string{"", "[]", "false", "0", "0s"}, f.Value.String()):
			break
		case f.Value.Type() == "bool":
			ret = append(ret, "--"+f.Name)
		case strings.HasSuffix(f.Value.Type(), "Slice"):
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				ret = append(ret, "--"+f.Name, strings.Join(sv.GetSlice(), ","))
			}
		case strings.HasSuffix(f.Value.Type(), "Array"):
			if sv, ok := f.Value.(pflag.SliceValue); ok {
				for _, v := range sv.GetSlice() {
					ret = append(ret, "--"+f.Name, v)
				}
			}
		default:
			ret = append(ret, "--"+f.Name, f.Value.String())
		}
	})

	return ret
}
