package presentation

// Exporter converts a Report into one or more output artifacts.
type Exporter interface {
	Export(report Report) ([]Output, error)
}
