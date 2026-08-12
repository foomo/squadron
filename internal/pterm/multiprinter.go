package pterm

import (
	"os"

	"github.com/pterm/pterm"
)

type MultiPrinter interface {
	NewSpinner(prefix string) Spinner
	Stop()
}

// multiPrinterHeightMargin reserves room for the Info/Success lines
// printed around each MultiPrinter block so the live area never grows
// taller than the terminal.
const multiPrinterHeightMargin = 4

func MustNewMultiPrinter(targetCount int) MultiPrinter {
	var (
		err   error
		value MultiPrinter
	)

	_, ci := os.LookupEnv("CI")

	if ci || targetCount > pterm.GetTerminalHeight()-multiPrinterHeightMargin {
		value, err = NewNoopMultiPrinter()
	} else {
		value, err = NewStandardMultiPrinter()
	}

	if err != nil {
		pterm.Fatal.Print(err)
	}

	return value
}
