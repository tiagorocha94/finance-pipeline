package presentation

import "tiagorocha94/household-finance-pipeline/internal/domain"

// Output represents a rendered artifact
// produced from a reporting generator.
type Output struct {
	Name string
	File domain.OutputFile
}
