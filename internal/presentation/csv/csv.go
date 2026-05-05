package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"strings"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
)

// Exporter writes one transactions CSV file per person, containing all of that
// person's expenses and incomes with no aggregation. Files are named after the
// person, e.g. Alice.csv, Bob.csv.
//
// Columns: Date, Person, Type, Amount, Currency, Category, Description
type Exporter struct{}

func New() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Export(r presentation.Report) ([]presentation.Output, error) {
	var outputs []presentation.Output

	for _, p := range r.People {
		content, err := buildCSV(p)
		if err != nil {
			return nil, fmt.Errorf("build csv for %s: %w", p.Name, err)
		}
		outputs = append(outputs, presentation.Output{
			Name: "csv",
			File: domain.OutputFile{
				Name:    strings.ToLower(p.Name) + ".csv",
				Content: content,
			},
		})
	}

	return outputs, nil
}

func buildCSV(p presentation.PersonView) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write([]string{"Date", "Person", "Type", "Amount", "Currency", "Category", "Description"}); err != nil {
		return nil, fmt.Errorf("write header: %w", err)
	}

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

	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("flush: %w", err)
	}

	return buf.Bytes(), nil
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
