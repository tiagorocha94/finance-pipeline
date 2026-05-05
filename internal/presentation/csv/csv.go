package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
)

// Exporter writes a flat transactions.csv containing every expense and income
// entry across all people, with no aggregation. The output is suitable for
// importing into spreadsheets or other tools.
//
// Columns: Date, Person, Type, Amount, Currency, Category, Description
type Exporter struct{}

func New() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Export(r presentation.Report) ([]presentation.Output, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{"Date", "Person", "Type", "Amount", "Currency", "Category", "Description"}); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}

	for _, p := range r.People {
		for _, ex := range p.Expenses {
			if err := w.Write(expenseRow(ex)); err != nil {
				return nil, fmt.Errorf("write expense row: %w", err)
			}
		}
		for _, in := range p.Incomes {
			if err := w.Write(incomeRow(in)); err != nil {
				return nil, fmt.Errorf("write income row: %w", err)
			}
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush csv: %w", err)
	}

	return []presentation.Output{
		{
			Name: "csv",
			File: domain.OutputFile{
				Name:    "transactions.csv",
				Content: buf.Bytes(),
			},
		},
	}, nil
}

// ── row builders ──────────────────────────────────────────

func expenseRow(e domain.Expense) []string {
	return []string{
		e.Date.Format("2006-01-02"),
		e.PersonName,
		"expense",
		fmt.Sprintf("%.2f", e.Amount.Value),
		e.Amount.Currency,
		e.Category,
		e.Note,
	}
}

func incomeRow(i domain.Income) []string {
	return []string{
		i.Date.Format("2006-01-02"),
		i.PersonName,
		"income",
		fmt.Sprintf("%.2f", i.Amount.Value),
		i.Amount.Currency,
		i.Source,
		i.Note,
	}
}
