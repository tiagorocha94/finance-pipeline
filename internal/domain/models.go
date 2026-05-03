package domain

import "time"

// Money is a simple value object for monetary amounts.
type Money struct {
	Value    float64
	Currency string
}

// Expense represents a single expense entry for a person.
type Expense struct {
	PersonName string
	Amount     Money
	Category   string
	Date       time.Time
	Note       string
}

// Income represents a single income entry for a person.
type Income struct {
	PersonName string
	Amount     Money
	Source     string
	Date       time.Time
	Note       string
}

// Person represents a household member.
type Person struct {
	Name string
}

// CategoryGroup holds the budget target and actual spending for one configured group.
type CategoryGroup struct {
	Name          string
	TargetPercent float64 // e.g. 0.50 for 50%
	TargetAmount  float64 // TargetPercent × total income
	ActualSpent   float64 // sum of expenses mapped to this group
	Delta         float64 // TargetAmount − ActualSpent (positive = under budget)
}

// PersonData is the parsed result of one person's file.
type PersonData struct {
	Person         Person
	Expenses       []Expense
	Incomes        []Income
	CategoryGroups []CategoryGroup
}

// HouseholdData is the merged view of all people in the household.
type HouseholdData struct {
	Persons        []Person
	Expenses       []Expense
	Incomes        []Income
	CategoryGroups []CategoryGroup
}
