package ingestion

import (
	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/domain"
)

type Ingestor interface {
	Ingest(cfg config.Config, month string) ([]domain.RawFile, error)
}
