package util

import (
	"bytes"
	"fmt"

	"github.com/alecthomas/chroma"
	"github.com/alecthomas/chroma/formatters"
	"github.com/alecthomas/chroma/lexers"
	"github.com/alecthomas/chroma/styles"
	"github.com/pterm/pterm"
)

func Highlight(source string) string {
	out := &numberWriter{
		w:           bytes.NewBufferString(""),
		currentLine: 1,
	}
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

	if err = f.Format(out, s, it); err != nil {
		pterm.Error.Println(err.Error())
	}

	return out.w.String()
}

type numberWriter struct {
	w           *bytes.Buffer
	currentLine uint64
	buf         []byte
}

func (w *numberWriter) Write(p []byte) (int, error) {
	// Early return.
	// Can't calculate the line numbers until the line breaks are made, so store them all in a buffer.
	if !bytes.Contains(p, []byte{'\n'}) {
		w.buf = append(w.buf, p...)
		return len(p), nil
	}

	var (
		original = p
		tokenLen uint
	)
	for i, c := range original {
		tokenLen++
		if c != '\n' {
			continue
		}

		token := p[:tokenLen]
		p = original[i+1:]
		tokenLen = 0

		format := "%4d |\t%s%s"
		if w.currentLine > 9999 {
			format = "%d |\t%s%s"
		}
		format = "\033[34m" + format + "\033[0m"

		if _, err := fmt.Fprintf(w.w, format, w.currentLine, string(w.buf), string(token)); err != nil {
			return i + 1, err
		}
		w.buf = w.buf[:0]
		w.currentLine++
	}

	if len(p) > 0 {
		w.buf = append(w.buf, p...)
	}
	return len(original), nil
}
