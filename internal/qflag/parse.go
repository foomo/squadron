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
			ret = append(ret, "--"+f.Name, strings.Join(f.Value.(pflag.SliceValue).GetSlice(), ","))
		case strings.HasSuffix(f.Value.Type(), "Array"):
			for _, v := range f.Value.(pflag.SliceValue).GetSlice() {
				ret = append(ret, "--"+f.Name, v)
			}
		default:
			ret = append(ret, "--"+f.Name, f.Value.String())
		}
	})
	return ret
}
