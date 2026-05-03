package xlsx

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"tiagorocha94/household-finance-pipeline/internal/domain"
)

const (
	sheetExpenses = "Despesas"
	sheetIncomes  = "Receita"

	// Column indices (0-based). Row 0 is the title, row 1 the header, data from row 2.
	colDate     = 0  // "Data e hora"
	colCategory = 1  // "Categoria"
	colAmount   = 3  // "Valor na moeda padrão"
	colCurrency = 4  // "Moeda padrão"
	colNote     = 10 // "Comentário"
)

// Parser implements parsing.Parser for XLSX files exported from the finance app.
// Each file corresponds to one person; the filename (without extension) is used
// as the person name, matching the convention of the CSV parser.
type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(file domain.RawFile) (domain.PersonData, error) {
	personName := extractPersonName(file.Name)

	f, err := excelize.OpenReader(bytes.NewReader(file.Content))
	if err != nil {
		return domain.PersonData{}, fmt.Errorf("open workbook: %w", err)
	}
	defer f.Close() //nolint: errcheck

	expenses, err := parseExpenses(f, personName)
	if err != nil {
		return domain.PersonData{}, fmt.Errorf("sheet %q: %w", sheetExpenses, err)
	}

	incomes, err := parseIncomes(f, personName)
	if err != nil {
		return domain.PersonData{}, fmt.Errorf("sheet %q: %w", sheetIncomes, err)
	}

	return domain.PersonData{
		Person:   domain.Person{Name: personName},
		Expenses: expenses,
		Incomes:  incomes,
	}, nil
}

func (p *Parser) Extension() string {
	return ".xlsx"
}

// -------------------- sheet parsers --------------------

func parseExpenses(f *excelize.File, personName string) ([]domain.Expense, error) {
	rows, err := sheetRows(f, sheetExpenses)
	if err != nil {
		return nil, err
	}

	var expenses []domain.Expense
	for i, row := range rows {
		money, date, note, err := parseRow(row, i)
		if err != nil {
			return nil, err
		}
		expenses = append(expenses, domain.Expense{
			PersonName: personName,
			Amount:     money,
			Category:   normaliseCategory(row[colCategory]),
			Date:       date,
			Note:       note,
		})
	}
	return expenses, nil
}

func parseIncomes(f *excelize.File, personName string) ([]domain.Income, error) {
	rows, err := sheetRows(f, sheetIncomes)
	if err != nil {
		return nil, err
	}

	var incomes []domain.Income
	for i, row := range rows {
		money, date, note, err := parseRow(row, i)
		if err != nil {
			return nil, err
		}
		incomes = append(incomes, domain.Income{
			PersonName: personName,
			Amount:     money,
			Source:     normaliseCategory(row[colCategory]),
			Date:       date,
			Note:       note,
		})
	}
	return incomes, nil
}

// -------------------- helpers --------------------

// sheetRows returns the data rows of a sheet, skipping the title row (index 0)
// and the header row (index 1).
func sheetRows(f *excelize.File, sheet string) ([][]string, error) {
	all, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}
	// Row 0: title; Row 1: column headers; Row 2+: data.
	if len(all) < 2 {
		return nil, nil
	}
	return all[2:], nil
}

// parseRow extracts money, date, and note from a data row.
func parseRow(row []string, idx int) (domain.Money, time.Time, string, error) {
	for len(row) <= colCurrency {
		row = append(row, "")
	}

	if strings.TrimSpace(row[colAmount]) == "" {
		return domain.Money{}, time.Time{}, "", fmt.Errorf("row %d: empty amount", idx+2)
	}

	amount, err := parseAmount(row[colAmount])
	if err != nil {
		return domain.Money{}, time.Time{}, "", fmt.Errorf("row %d: parse amount %q: %w", idx+2, row[colAmount], err)
	}

	date, err := parseDate(row[colDate])
	if err != nil {
		return domain.Money{}, time.Time{}, "", fmt.Errorf("row %d: parse date %q: %w", idx+2, row[colDate], err)
	}

	note := ""
	if len(row) > colNote {
		note = strings.TrimSpace(row[colNote])
	}

	return domain.Money{
		Value:    amount,
		Currency: strings.TrimSpace(row[colCurrency]),
	}, date, note, nil
}

// parseAmount converts strings like "5.6" or "2 700,45" to float64.
func parseAmount(s string) (float64, error) {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\u00a0", "")
	s = strings.ReplaceAll(s, ",", ".")

	var v float64
	_, err := fmt.Sscanf(s, "%f", &v)
	return v, err
}

// parseDate tries common date formats produced by the finance app.
// Only 4-digit year formats are included to avoid ambiguity.
var dateLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02",
	"1/2/2006",
	"01/02/2006",
	"2/1/2006", // day/month/year variant
	"02/01/2006",
	"01-02-06",
}

func parseDate(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognised date format %q", s)
}

// normaliseCategory trims whitespace and title-cases the first letter so that
// "seguros" and "Seguros" both become "Seguros". This prevents silent mismatches
// against category_groups config keys.
func normaliseCategory(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func extractPersonName(filename string) string {
	base := filepath.Base(filename)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
