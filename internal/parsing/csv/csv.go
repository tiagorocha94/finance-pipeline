package csv

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"tiagorocha94/household-finance-pipeline/internal/domain"
)

// Parser implements parsing.Parser for CSV files.
type Parser struct{}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(file domain.RawFile) (domain.PersonData, error) {
	personName := extractPersonName(file.Name)

	rows, err := readRows(file.Content)
	if err != nil {
		return domain.PersonData{}, err
	}

	expenses, incomes, err := parseRows(rows, personName)
	if err != nil {
		return domain.PersonData{}, err
	}

	return domain.PersonData{
		Person: domain.Person{
			Name: personName,
		},
		Expenses: expenses,
		Incomes:  incomes,
	}, nil
}

func (p *Parser) Extension() string {
	return ".csv"
}

// -------------------- helpers --------------------

func extractPersonName(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

func readRows(content []byte) ([][]string, error) {
	r := csv.NewReader(bytes.NewReader(content))
	return r.ReadAll()
}

func parseRows(rows [][]string, personName string) ([]domain.Expense, []domain.Income, error) {
	var expenses []domain.Expense
	var incomes []domain.Income

	for i, row := range rows {
		if i == 0 {
			continue // header
		}

		if len(row) < 6 {
			return nil, nil, fmt.Errorf("invalid row: %v", row)
		}

		money, date, err := parseCommon(row)
		if err != nil {
			return nil, nil, err
		}

		switch row[0] {
		case "expense":
			expenses = append(expenses, domain.Expense{
				PersonName: personName,
				Amount:     money,
				Category:   row[3],
				Date:       date,
				Note:       row[5],
			})

		case "income":
			incomes = append(incomes, domain.Income{
				PersonName: personName,
				Amount:     money,
				Source:     row[3],
				Date:       date,
				Note:       row[5],
			})

		default:
			return nil, nil, fmt.Errorf("unknown type: %s", row[0])
		}
	}

	return expenses, incomes, nil
}

func parseCommon(row []string) (domain.Money, time.Time, error) {
	amount, err := strconv.ParseFloat(row[1], 64)
	if err != nil {
		return domain.Money{}, time.Time{}, err
	}

	date, err := time.Parse("2006-01-02", row[4])
	if err != nil {
		return domain.Money{}, time.Time{}, err
	}

	return domain.Money{
		Value:    amount,
		Currency: row[2],
	}, date, nil
}
