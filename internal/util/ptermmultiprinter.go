package util

import (
	"os"

	"github.com/pterm/pterm"
)

type PTermMultiPrinter struct {
	printer *pterm.MultiPrinter
}

func MustNewPTermMultiPrinter() *PTermMultiPrinter {
	printer, err := NewPTermMultiPrinter()
	if err != nil {
		pterm.Fatal.Println(err)
	}
	return printer
}

func NewPTermMultiPrinter() (*PTermMultiPrinter, error) {
	printer, err := pterm.DefaultMultiPrinter.WithWriter(os.Stdout).Start()
	if err != nil {
		return nil, err
	}
	return &PTermMultiPrinter{printer: printer}, nil
}

func (s *PTermMultiPrinter) NewSpinner(prefix string) *PTermSpinner {
	return NewPTermSpinner(s.printer.NewWriter(), prefix)
}

func (s *PTermMultiPrinter) Stop() {
	if s.printer.IsActive {
		if _, err := s.printer.Stop(); err != nil {
			pterm.Fatal.Println(err)
		}
	}
}
