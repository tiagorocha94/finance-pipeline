package parsing

import "tiagorocha94/household-finance-pipeline/internal/domain"

// Parser converts a RawFile into structured PersonData.
type Parser interface {
	Parse(file domain.RawFile) (domain.PersonData, error)
}
