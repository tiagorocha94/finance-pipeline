package aggregation

import "tiagorocha94/household-finance-pipeline/internal/domain"

type Aggregator interface {
	Aggregate(data []domain.PersonData) (domain.HouseholdData, []domain.PersonData, error)
}
