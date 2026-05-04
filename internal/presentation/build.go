package presentation

import (
	"sort"

	"tiagorocha94/household-finance-pipeline/internal/domain"
)

// BuildReport converts aggregated household data and per-person data into the
// presentation model. All output slices are sorted for deterministic ordering:
// people alphabetically, categories/sources by amount descending, raw
// transactions by date descending.
func BuildReport(h domain.HouseholdData, persons []domain.PersonData) Report {
	householdExpense := map[string]float64{}
	householdIncome := map[string]float64{}
	personExpense := map[string]map[string]float64{}
	personIncome := map[string]map[string]float64{}

	var totalIncome, totalExpense float64

	for _, e := range h.Expenses {
		totalExpense += e.Amount.Value
		householdExpense[e.Category] += e.Amount.Value
		if personExpense[e.PersonName] == nil {
			personExpense[e.PersonName] = map[string]float64{}
		}
		personExpense[e.PersonName][e.Category] += e.Amount.Value
	}

	for _, i := range h.Incomes {
		totalIncome += i.Amount.Value
		householdIncome[i.Source] += i.Amount.Value
		if personIncome[i.PersonName] == nil {
			personIncome[i.PersonName] = map[string]float64{}
		}
		personIncome[i.PersonName][i.Source] += i.Amount.Value
	}

	household := HouseholdView{
		Totals: Totals{
			Income:  domain.Money{Value: totalIncome, Currency: "EUR"},
			Expense: domain.Money{Value: totalExpense, Currency: "EUR"},
			Balance: domain.Money{Value: totalIncome - totalExpense, Currency: "EUR"},
		},
		CategoryGroups: h.CategoryGroups,
	}

	for k, v := range householdExpense {
		household.ExpensesByCategory = append(household.ExpensesByCategory, CategoryTotal{
			Category: k,
			Total:    domain.Money{Value: v, Currency: "EUR"},
		})
	}
	sort.Slice(household.ExpensesByCategory, func(i, j int) bool {
		return household.ExpensesByCategory[i].Total.Value > household.ExpensesByCategory[j].Total.Value
	})

	for k, v := range householdIncome {
		household.IncomeBySource = append(household.IncomeBySource, SourceTotal{
			Source: k,
			Total:  domain.Money{Value: v, Currency: "EUR"},
		})
	}
	sort.Slice(household.IncomeBySource, func(i, j int) bool {
		return household.IncomeBySource[i].Total.Value > household.IncomeBySource[j].Total.Value
	})

	// Build name → PersonData lookup for groups and raw transactions.
	personGroups := make(map[string][]domain.CategoryGroup, len(persons))
	personExpenses := make(map[string][]domain.Expense, len(persons))
	personIncomes := make(map[string][]domain.Income, len(persons))
	for _, pd := range persons {
		personGroups[pd.Person.Name] = pd.CategoryGroups
		personExpenses[pd.Person.Name] = pd.Expenses
		personIncomes[pd.Person.Name] = pd.Incomes
	}

	// Collect all person names and sort alphabetically for deterministic output.
	allNames := make(map[string]bool)
	for name := range personExpense {
		allNames[name] = true
	}
	for name := range personIncome {
		allNames[name] = true
	}
	sortedNames := make([]string, 0, len(allNames))
	for name := range allNames {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	var people []PersonView
	for _, name := range sortedNames {
		pv := buildPersonView(
			name,
			personExpense[name],
			personIncome[name],
			personGroups[name],
			personExpenses[name],
			personIncomes[name],
		)
		people = append(people, pv)
	}

	return Report{
		Household: household,
		People:    people,
	}
}

func buildPersonView(
	name string,
	expenses map[string]float64,
	incomes map[string]float64,
	groups []domain.CategoryGroup,
	rawExpenses []domain.Expense,
	rawIncomes []domain.Income,
) PersonView {
	pv := PersonView{
		Name:           name,
		CategoryGroups: groups,
		Expenses:       rawExpenses,
		Incomes:        rawIncomes,
	}

	var pIncome, pExpense float64
	for _, v := range expenses {
		pExpense += v
	}
	for _, v := range incomes {
		pIncome += v
	}

	pv.Totals = Totals{
		Income:  domain.Money{Value: pIncome, Currency: "EUR"},
		Expense: domain.Money{Value: pExpense, Currency: "EUR"},
		Balance: domain.Money{Value: pIncome - pExpense, Currency: "EUR"},
	}

	for k, v := range expenses {
		pv.ExpensesByCategory = append(pv.ExpensesByCategory, CategoryTotal{
			Category: k,
			Total:    domain.Money{Value: v, Currency: "EUR"},
		})
	}
	sort.Slice(pv.ExpensesByCategory, func(i, j int) bool {
		return pv.ExpensesByCategory[i].Total.Value > pv.ExpensesByCategory[j].Total.Value
	})

	for k, v := range incomes {
		pv.IncomeBySource = append(pv.IncomeBySource, SourceTotal{
			Source: k,
			Total:  domain.Money{Value: v, Currency: "EUR"},
		})
	}
	sort.Slice(pv.IncomeBySource, func(i, j int) bool {
		return pv.IncomeBySource[i].Total.Value > pv.IncomeBySource[j].Total.Value
	})

	// Sort raw transactions by date descending for consistent JSON output.
	sort.Slice(pv.Expenses, func(i, j int) bool {
		return pv.Expenses[i].Date.After(pv.Expenses[j].Date)
	})
	sort.Slice(pv.Incomes, func(i, j int) bool {
		return pv.Incomes[i].Date.After(pv.Incomes[j].Date)
	})

	return pv
}
