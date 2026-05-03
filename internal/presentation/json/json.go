package json

import (
	"encoding/json"
	"fmt"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
)

type Exporter struct{}

func New() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Export(r presentation.Report) ([]presentation.Output, error) {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal report: %w", err)
	}

	return []presentation.Output{
		{
			Name: "json",
			File: domain.OutputFile{
				Name:    "data.json",
				Content: b,
			},
		},
	}, nil
}