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

func findOutput(outputs []presentation.Output, name string) *presentation.Output {
	for i := range outputs {
		if outputs[i].File.Name == name {
			return &outputs[i]
		}
	}
	return nil
}

// ── Output count and filenames ────────────────────────────

func TestExport_OneFilePerPerson(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 2 {
		t.Fatalf("want 2 outputs (one per person), got %d", len(outputs))
	}
}

func TestExport_Filenames(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if findOutput(outputs, "alice.csv") == nil {
		t.Error("want alice.csv in outputs")
	}
	if findOutput(outputs, "bob.csv") == nil {
		t.Error("want bob.csv in outputs")
	}
}

// ── Headers ───────────────────────────────────────────────

func TestExport_Header(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"Date", "Person", "Type", "Amount", "Currency", "Category", "Description"}
	for _, o := range outputs {
		records := parseCSV(t, o.File.Content)
		if len(records) == 0 {
			t.Fatalf("%s: no records", o.File.Name)
		}
		for i, col := range want {
			if records[0][i] != col {
				t.Errorf("%s header[%d]: want %q, got %q", o.File.Name, i, col, records[0][i])
			}
		}
	}
}

// ── Row counts per person ─────────────────────────────────

func TestExport_AliceRowCount(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	o := findOutput(outputs, "alice.csv")
	if o == nil {
		t.Fatal("alice.csv not found")
	}
	records := parseCSV(t, o.File.Content)
	// 1 header + 2 expenses + 1 income = 4
	if len(records) != 4 {
		t.Errorf("alice.csv: want 4 rows, got %d", len(records))
	}
}

func TestExport_BobRowCount(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	o := findOutput(outputs, "bob.csv")
	if o == nil {
		t.Fatal("bob.csv not found")
	}
	records := parseCSV(t, o.File.Content)
	// 1 header + 1 expense = 2
	if len(records) != 2 {
		t.Errorf("bob.csv: want 2 rows, got %d", len(records))
	}
}

// ── Row content ───────────────────────────────────────────

func TestExport_ExpenseRow(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	o := findOutput(outputs, "alice.csv")
	if o == nil {
		t.Fatal("alice.csv not found")
	}
	records := parseCSV(t, o.File.Content)
	row := records[1] // first data row
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

	o := findOutput(outputs, "alice.csv")
	if o == nil {
		t.Fatal("alice.csv not found")
	}
	records := parseCSV(t, o.File.Content)
	row := records[3] // after 2 expenses
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

func TestExport_IsolationBetweenPeople(t *testing.T) {
	// Bob's file must not contain Alice's rows.
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	o := findOutput(outputs, "bob.csv")
	if o == nil {
		t.Fatal("bob.csv not found")
	}
	records := parseCSV(t, o.File.Content)
	for _, row := range records[1:] {
		if row[1] != "Bob" {
			t.Errorf("bob.csv contains row for %q, want Bob only", row[1])
		}
	}
}

// ── Edge cases ────────────────────────────────────────────

func TestExport_EmptyNote(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(makeReport())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	o := findOutput(outputs, "alice.csv")
	if o == nil {
		t.Fatal("alice.csv not found")
	}
	records := parseCSV(t, o.File.Content)
	// Row 2 (index 2): Alice's transport expense — no note
	if records[2][6] != "" {
		t.Errorf("empty note: want empty string, got %q", records[2][6])
	}
}

func TestExport_EmptyReport(t *testing.T) {
	e := csvexp.New()
	outputs, err := e.Export(presentation.Report{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(outputs) != 0 {
		t.Errorf("empty report: want 0 outputs, got %d", len(outputs))
	}
}

func TestExport_CommaInNote(t *testing.T) {
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
