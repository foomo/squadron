package pterm

import (
	"os"

	"github.com/pterm/pterm"
)

type StandardMultiPrinter struct {
	printer *pterm.MultiPrinter
}

func NewStandardMultiPrinter() (*StandardMultiPrinter, error) {
	printer, err := pterm.DefaultMultiPrinter.WithWriter(os.Stdout).Start()
	if err != nil {
		return nil, err
	}

	return &StandardMultiPrinter{printer: printer}, nil
}

func (s *StandardMultiPrinter) NewSpinner(prefix string) Spinner {
	return NewStandardSpinner(s.printer.NewWriter(), prefix)
}

func (s *StandardMultiPrinter) Stop() {
	if s.printer.IsActive {
		if _, err := s.printer.Stop(); err != nil {
			pterm.Fatal.Println(err)
		}
	}
}
