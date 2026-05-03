package presentation

import "tiagorocha94/household-finance-pipeline/internal/domain"

// Report is the full analytical model for presentation layer.
type Report struct {
	Household HouseholdView
	People    []PersonView
}

// HouseholdView is aggregated household-level data.
type HouseholdView struct {
	Totals Totals

	ExpensesByCategory []CategoryTotal
	IncomeBySource     []SourceTotal
	CategoryGroups     []domain.CategoryGroup // populated when budget categories are configured
}

// PersonView is per-person breakdown.
type PersonView struct {
	Name string

	Totals Totals

	ExpensesByCategory []CategoryTotal
	IncomeBySource     []SourceTotal
	CategoryGroups     []domain.CategoryGroup // populated when budget categories are configured

	Expenses []domain.Expense // raw transactions for the month
	Incomes  []domain.Income  // raw transactions for the month
}

type Totals struct {
	Income  domain.Money
	Expense domain.Money
	Balance domain.Money
}

type CategoryTotal struct {
	Category string
	Total    domain.Money
}

type SourceTotal struct {
	Source string
	Total  domain.Money
}
