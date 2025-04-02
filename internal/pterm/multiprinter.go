package pterm

import (
	"os"

	"github.com/pterm/pterm"
)

type MultiPrinter interface {
	NewSpinner(prefix string) Spinner
	Stop()
}

func MustNewMultiPrinter() MultiPrinter {
	var err error
	var value MultiPrinter
	if os.Getenv("CI") == "true" {
		value, err = NewNoopMultiPrinter()
	} else {
		value, err = NewStandardMultiPrinter()
	}
	if err != nil {
		pterm.Fatal.Print(err)
	}
	return value
}
