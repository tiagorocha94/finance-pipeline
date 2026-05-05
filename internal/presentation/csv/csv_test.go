package csv_test

import (
	"bytes"
	"encoding/csv"
	"testing"
	"time"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
	csvexp "tiagorocha94/household-finance-pipeline/internal/presentation/csv"
)

func date(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func makeReport() presentation.Report {
	return presentation.Report{
		People: []presentation.PersonView{
			{
				Name: "Alice",
				Expenses: []domain.Expense{
					{PersonName: "Alice", Date: date(2026, 4, 1), Amount: domain.Money{Value: 42.50, Currency: "EUR"}, Category: "Food", Note: "Supermarket"},
					{PersonName: "Alice", Date: date(2026, 4, 5), Amount: domain.Money{Value: 9.99, Currency: "EUR"}, Category: "Transport", Note: ""},
				},
				Incomes: []domain.Income{
					{PersonName: "Alice", Date: date(2026, 4, 1), Amount: domain.Money{Value: 2500.00, Currency: "EUR"}, Source: "Salary", Note: "April"},
				},
			},
			{
				Name: "Bob",
				Expenses: []domain.Expense{
					{PersonName: "Bob", Date: date(2026, 4, 3), Amount: domain.Money{Value: 100.00, Currency: "EUR"}, Category: "Restaurant", Note: "Dinner at X"},
				},
				Incomes: []domain.Income{},
			},
		},
	}
}

func parseCSV(t *testing.T, content []byte) [][]string {
	t.Helper()
	r := csv.NewReader(bytes.NewReader(content))
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse csv output: %v", err)
	}
	return records
}

func TestExport_Header(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, outputs[0].File.Content)
	want := []string{"Date", "Person", "Type", "Amount", "Currency", "Category", "Description"}
	if len(records) == 0 {
		t.Fatal("no records in output")
	}
	for i, col := range want {
		if records[0][i] != col {
			t.Errorf("header[%d]: want %q, got %q", i, col, records[0][i])
		}
	}
}

func TestExport_RowCount(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, outputs[0].File.Content)
	// 1 header + 2 Alice expenses + 1 Alice income + 1 Bob expense = 5
	if len(records) != 5 {
		t.Errorf("want 5 rows (header + 4 data), got %d", len(records))
	}
}

func TestExport_ExpenseRow(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, outputs[0].File.Content)
	// First data row: Alice's first expense
	row := records[1]
	checks := map[int]string{
		0: "2026-04-01",
		1: "Alice",
		2: "expense",
		3: "42.50",
		4: "EUR",
		5: "Food",
		6: "Supermarket",
	}
	for col, want := range checks {
		if row[col] != want {
			t.Errorf("expense row col %d: want %q, got %q", col, want, row[col])
		}
	}
}

func TestExport_IncomeRow(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, outputs[0].File.Content)
	// Third data row (index 3): Alice's income
	row := records[3]
	checks := map[int]string{
		0: "2026-04-01",
		1: "Alice",
		2: "income",
		3: "2500.00",
		4: "EUR",
		5: "Salary",
		6: "April",
	}
	for col, want := range checks {
		if row[col] != want {
			t.Errorf("income row col %d: want %q, got %q", col, want, row[col])
		}
	}
}

func TestExport_EmptyNote(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	records := parseCSV(t, outputs[0].File.Content)
	// Second data row (index 2): Alice's transport expense with empty note
	if records[2][6] != "" {
		t.Errorf("empty note: want empty string, got %q", records[2][6])
	}
}

func TestExport_OutputFilename(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 1 {
		t.Fatalf("want 1 output, got %d", len(outputs))
	}
	if outputs[0].File.Name != "transactions.csv" {
		t.Errorf("filename: want transactions.csv, got %q", outputs[0].File.Name)
	}
}

func TestExport_EmptyReport(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(presentation.Report{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	records := parseCSV(t, outputs[0].File.Content)
	// Only the header row
	if len(records) != 1 {
		t.Errorf("empty report: want 1 row (header only), got %d", len(records))
	}
}

func TestExport_ValidCSV(t *testing.T) {
	// Ensure values with commas are properly quoted by encoding/csv.
	e := csvexp.New()
	r := presentation.Report{
		People: []presentation.PersonView{
			{
				Name: "Alice",
				Expenses: []domain.Expense{
					{PersonName: "Alice", Date: date(2026, 1, 1), Amount: domain.Money{Value: 10, Currency: "EUR"}, Category: "Food", Note: "Cafe, downtown"},
				},
			},
		},
	}
	outputs, err := e.Export(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	records := parseCSV(t, outputs[0].File.Content)
	if records[1][6] != "Cafe, downtown" {
		t.Errorf("comma in note: want %q, got %q", "Cafe, downtown", records[1][6])
	}
}
