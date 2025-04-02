package pterm

type NoopMultiPrinter struct {
}

func NewNoopMultiPrinter() (*NoopMultiPrinter, error) {
	return &NoopMultiPrinter{}, nil
}

func (s *NoopMultiPrinter) NewSpinner(prefix string) Spinner {
	return NewNoopSpinner(prefix)
}

func (s *NoopMultiPrinter) Stop() {}
