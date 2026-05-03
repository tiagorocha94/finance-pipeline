package csv_test

import (
	"testing"
	"time"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/parsing/csv"
)

func file(name, content string) domain.RawFile {
	return domain.RawFile{Name: name, Content: []byte(content)}
}

func TestParse_ExpenseAndIncome(t *testing.T) {
	p := csv.New()
	raw := file("Alice.csv", `type,amount,currency,category,date,note
expense,120.50,EUR,Food,2026-01-15,Supermarket
income,2500.00,EUR,Salary,2026-01-01,January salary
expense,45.00,EUR,Transport,2026-01-10,Bus pass
`)

	data, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if data.Person.Name != "Alice" {
		t.Errorf("person name: want Alice, got %q", data.Person.Name)
	}
	if len(data.Expenses) != 2 {
		t.Fatalf("expenses: want 2, got %d", len(data.Expenses))
	}
	if len(data.Incomes) != 1 {
		t.Fatalf("incomes: want 1, got %d", len(data.Incomes))
	}

	e := data.Expenses[0]
	if e.Amount.Value != 120.50 {
		t.Errorf("expense amount: want 120.50, got %.2f", e.Amount.Value)
	}
	if e.Category != "Food" {
		t.Errorf("expense category: want Food, got %q", e.Category)
	}
	if e.Note != "Supermarket" {
		t.Errorf("expense note: want Supermarket, got %q", e.Note)
	}
	wantDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	if !e.Date.Equal(wantDate) {
		t.Errorf("expense date: want %v, got %v", wantDate, e.Date)
	}

	inc := data.Incomes[0]
	if inc.Amount.Value != 2500.00 {
		t.Errorf("income amount: want 2500.00, got %.2f", inc.Amount.Value)
	}
	if inc.Source != "Salary" {
		t.Errorf("income source: want Salary, got %q", inc.Source)
	}
}

func TestParse_PersonNameFromFilename(t *testing.T) {
	p := csv.New()
	tests := []struct {
		filename string
		want     string
	}{
		{"Bob.csv", "Bob"},
		{"path/to/Charlie.csv", "Charlie"},
		{"diana-smith.csv", "diana-smith"},
	}

	header := "type,amount,currency,category,date,note\n"
	for _, tt := range tests {
		data, err := p.Parse(file(tt.filename, header))
		if err != nil {
			t.Fatalf("%s: unexpected error: %v", tt.filename, err)
		}
		if data.Person.Name != tt.want {
			t.Errorf("%s: want %q, got %q", tt.filename, tt.want, data.Person.Name)
		}
	}
}

func TestParse_EmptyFile(t *testing.T) {
	p := csv.New()
	data, err := p.Parse(file("Alice.csv", "type,amount,currency,category,date,note\n"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(data.Expenses) != 0 || len(data.Incomes) != 0 {
		t.Error("expected no transactions for header-only file")
	}
}

func TestParse_UnknownType(t *testing.T) {
	p := csv.New()
	raw := file("Alice.csv", "type,amount,currency,category,date,note\nbogus,10,EUR,X,2026-01-01,note\n")
	_, err := p.Parse(raw)
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestParse_TooFewColumns(t *testing.T) {
	p := csv.New()
	raw := file("Alice.csv", "type,amount,currency\nexpense,10,EUR\n")
	_, err := p.Parse(raw)
	if err == nil {
		t.Error("expected error for row with too few columns")
	}
}

func TestParse_InvalidAmount(t *testing.T) {
	p := csv.New()
	raw := file("Alice.csv", "type,amount,currency,category,date,note\nexpense,not-a-number,EUR,Food,2026-01-01,\n")
	_, err := p.Parse(raw)
	if err == nil {
		t.Error("expected error for invalid amount")
	}
}

func TestParse_InvalidDate(t *testing.T) {
	p := csv.New()
	raw := file("Alice.csv", "type,amount,currency,category,date,note\nexpense,10.00,EUR,Food,not-a-date,\n")
	_, err := p.Parse(raw)
	if err == nil {
		t.Error("expected error for invalid date")
	}
}

func TestParse_PersonNameOnTransactions(t *testing.T) {
	p := csv.New()
	raw := file("Alice.csv", "type,amount,currency,category,date,note\nexpense,10,EUR,Food,2026-01-01,\n")
	data, err := p.Parse(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Expenses[0].PersonName != "Alice" {
		t.Errorf("PersonName on expense: want Alice, got %q", data.Expenses[0].PersonName)
	}
}

func TestCSVParser_Extension(t *testing.T) {
	p := csv.New()
	if ext := p.Extension(); ext != ".csv" {
		t.Errorf("Extension: want .csv, got %q", ext)
	}
}
