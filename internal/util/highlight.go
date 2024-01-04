package util

import (
	"bytes"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/pterm/pterm"
)

func Highlight(source string) string {
	var out bytes.Buffer
	// Determine lexer.
	l := lexers.Get("yaml")
	if l == nil {
		l = lexers.Analyse(source)
	}
	if l == nil {
		l = lexers.Fallback
	}
	l = chroma.Coalesce(l)

	// Determine formatter.
	f := formatters.Get("terminal256")
	if f == nil {
		f = formatters.Fallback
	}

	// Determine style.
	s := styles.Get("monokai")
	if s == nil {
		s = styles.Fallback
	}

	it, err := l.Tokenise(nil, source)
	if err != nil {
		pterm.Error.Println(err.Error())
	}
	err = f.Format(&out, s, it)
	if err != nil {
		pterm.Error.Println(err.Error())
	}

	return out.String()
}
