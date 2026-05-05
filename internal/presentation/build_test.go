package presentation_test

import (
	"testing"
	"time"

	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
)

// ── helpers ───────────────────────────────────────────────

func d(year, month, day int) time.Time {
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

func money(v float64) domain.Money {
	return domain.Money{Value: v, Currency: "EUR"}
}

func expense(person, category string, amount float64, date time.Time, note string) domain.Expense {
	return domain.Expense{
		PersonName: person,
		Category:   category,
		Amount:     money(amount),
		Date:       date,
		Note:       note,
	}
}

func income(person, source string, amount float64, date time.Time, note string) domain.Income {
	return domain.Income{
		PersonName: person,
		Source:     source,
		Amount:     money(amount),
		Date:       date,
		Note:       note,
	}
}

// ── Household totals ──────────────────────────────────────

func TestBuildReport_HouseholdTotals(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Alice", "Food", 200, d(2026, 1, 1), ""),
			expense("Bob", "Rent", 800, d(2026, 1, 1), ""),
		},
		Incomes: []domain.Income{
			income("Alice", "Salary", 2000, d(2026, 1, 1), ""),
			income("Bob", "Salary", 1500, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)

	if r.Household.Totals.Income.Value != 3500 {
		t.Errorf("total income: want 3500, got %.2f", r.Household.Totals.Income.Value)
	}
	if r.Household.Totals.Expense.Value != 1000 {
		t.Errorf("total expense: want 1000, got %.2f", r.Household.Totals.Expense.Value)
	}
	if r.Household.Totals.Balance.Value != 2500 {
		t.Errorf("balance: want 2500, got %.2f", r.Household.Totals.Balance.Value)
	}
}

func TestBuildReport_HouseholdCurrency(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{expense("Alice", "Food", 100, d(2026, 1, 1), "")},
		Incomes:  []domain.Income{income("Alice", "Salary", 500, d(2026, 1, 1), "")},
	}

	r := presentation.BuildReport(h, nil)

	for _, m := range []domain.Money{
		r.Household.Totals.Income,
		r.Household.Totals.Expense,
		r.Household.Totals.Balance,
	} {
		if m.Currency != "EUR" {
			t.Errorf("currency: want EUR, got %q", m.Currency)
		}
	}
}

// ── Household category aggregation ───────────────────────

func TestBuildReport_ExpensesByCategory_Aggregated(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Alice", "Food", 100, d(2026, 1, 1), ""),
			expense("Bob", "Food", 50, d(2026, 1, 2), ""),
			expense("Alice", "Rent", 800, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)

	cats := make(map[string]float64)
	for _, c := range r.Household.ExpensesByCategory {
		cats[c.Category] = c.Total.Value
	}

	if cats["Food"] != 150 {
		t.Errorf("Food: want 150, got %.2f", cats["Food"])
	}
	if cats["Rent"] != 800 {
		t.Errorf("Rent: want 800, got %.2f", cats["Rent"])
	}
}

func TestBuildReport_ExpensesByCategory_SortedDescending(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Alice", "Food", 100, d(2026, 1, 1), ""),
			expense("Alice", "Rent", 800, d(2026, 1, 1), ""),
			expense("Alice", "Transport", 50, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)
	cats := r.Household.ExpensesByCategory

	for i := 1; i < len(cats); i++ {
		if cats[i].Total.Value > cats[i-1].Total.Value {
			t.Errorf("ExpensesByCategory not sorted desc at index %d: %.2f > %.2f",
				i, cats[i].Total.Value, cats[i-1].Total.Value)
		}
	}
}

func TestBuildReport_IncomeBySource_Aggregated(t *testing.T) {
	h := domain.HouseholdData{
		Incomes: []domain.Income{
			income("Alice", "Salary", 2000, d(2026, 1, 1), ""),
			income("Bob", "Salary", 1500, d(2026, 1, 1), ""),
			income("Alice", "Freelance", 300, d(2026, 1, 5), ""),
		},
	}

	r := presentation.BuildReport(h, nil)

	srcs := make(map[string]float64)
	for _, s := range r.Household.IncomeBySource {
		srcs[s.Source] = s.Total.Value
	}

	if srcs["Salary"] != 3500 {
		t.Errorf("Salary: want 3500, got %.2f", srcs["Salary"])
	}
	if srcs["Freelance"] != 300 {
		t.Errorf("Freelance: want 300, got %.2f", srcs["Freelance"])
	}
}

func TestBuildReport_IncomeBySource_SortedDescending(t *testing.T) {
	h := domain.HouseholdData{
		Incomes: []domain.Income{
			income("Alice", "Freelance", 300, d(2026, 1, 1), ""),
			income("Alice", "Salary", 2000, d(2026, 1, 1), ""),
			income("Alice", "Bonus", 500, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)
	srcs := r.Household.IncomeBySource

	for i := 1; i < len(srcs); i++ {
		if srcs[i].Total.Value > srcs[i-1].Total.Value {
			t.Errorf("IncomeBySource not sorted desc at index %d", i)
		}
	}
}

// ── CategoryGroups passthrough ────────────────────────────

func TestBuildReport_CategoryGroupsPassedThrough(t *testing.T) {
	groups := []domain.CategoryGroup{
		{Name: "Essentials", TargetPercent: 0.5, TargetAmount: 1000, ActualSpent: 800, Delta: 200},
	}
	h := domain.HouseholdData{CategoryGroups: groups}

	r := presentation.BuildReport(h, nil)

	if len(r.Household.CategoryGroups) != 1 {
		t.Fatalf("want 1 category group, got %d", len(r.Household.CategoryGroups))
	}
	if r.Household.CategoryGroups[0].Name != "Essentials" {
		t.Errorf("group name: want Essentials, got %q", r.Household.CategoryGroups[0].Name)
	}
}

// ── People ordering ───────────────────────────────────────

func TestBuildReport_PeopleSortedAlphabetically(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Zara", "Food", 100, d(2026, 1, 1), ""),
			expense("Alice", "Food", 200, d(2026, 1, 1), ""),
			expense("Bob", "Food", 150, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)

	want := []string{"Alice", "Bob", "Zara"}
	for i, name := range want {
		if r.People[i].Name != name {
			t.Errorf("people[%d]: want %q, got %q", i, name, r.People[i].Name)
		}
	}
}

func TestBuildReport_IncludesPersonWithIncomeOnly(t *testing.T) {
	// Person with no expenses should still appear.
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Alice", "Food", 100, d(2026, 1, 1), ""),
		},
		Incomes: []domain.Income{
			income("Alice", "Salary", 2000, d(2026, 1, 1), ""),
			income("Bob", "Salary", 1500, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)

	names := make(map[string]bool)
	for _, p := range r.People {
		names[p.Name] = true
	}
	if !names["Bob"] {
		t.Error("Bob (income-only) should appear in People")
	}
}

// ── Person totals ─────────────────────────────────────────

func TestBuildReport_PersonTotals(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Alice", "Food", 200, d(2026, 1, 1), ""),
			expense("Alice", "Rent", 800, d(2026, 1, 2), ""),
		},
		Incomes: []domain.Income{
			income("Alice", "Salary", 3000, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)

	if len(r.People) != 1 {
		t.Fatalf("want 1 person, got %d", len(r.People))
	}
	p := r.People[0]

	if p.Totals.Expense.Value != 1000 {
		t.Errorf("person expense: want 1000, got %.2f", p.Totals.Expense.Value)
	}
	if p.Totals.Income.Value != 3000 {
		t.Errorf("person income: want 3000, got %.2f", p.Totals.Income.Value)
	}
	if p.Totals.Balance.Value != 2000 {
		t.Errorf("person balance: want 2000, got %.2f", p.Totals.Balance.Value)
	}
}

func TestBuildReport_PersonExpensesByCategory_SortedDescending(t *testing.T) {
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Alice", "Food", 100, d(2026, 1, 1), ""),
			expense("Alice", "Rent", 800, d(2026, 1, 1), ""),
			expense("Alice", "Transport", 50, d(2026, 1, 1), ""),
		},
	}

	r := presentation.BuildReport(h, nil)
	cats := r.People[0].ExpensesByCategory

	for i := 1; i < len(cats); i++ {
		if cats[i].Total.Value > cats[i-1].Total.Value {
			t.Errorf("person ExpensesByCategory not sorted desc at index %d", i)
		}
	}
	if cats[0].Category != "Rent" {
		t.Errorf("first category: want Rent, got %q", cats[0].Category)
	}
}

// ── Raw transactions on PersonView ───────────────────────

func TestBuildReport_PersonRawExpenses(t *testing.T) {
	rawExpenses := []domain.Expense{
		expense("Alice", "Food", 100, d(2026, 1, 3), "Late"),
		expense("Alice", "Food", 200, d(2026, 1, 1), "Early"),
	}
	persons := []domain.PersonData{
		{Person: domain.Person{Name: "Alice"}, Expenses: rawExpenses},
	}
	h := domain.HouseholdData{
		Expenses: rawExpenses,
	}

	r := presentation.BuildReport(h, persons)

	if len(r.People) == 0 {
		t.Fatal("no people in report")
	}
	p := r.People[0]

	if len(p.Expenses) != 2 {
		t.Fatalf("raw expenses: want 2, got %d", len(p.Expenses))
	}
	// Should be sorted date descending: Jan 3 first
	if !p.Expenses[0].Date.Equal(d(2026, 1, 3)) {
		t.Errorf("first expense date: want 2026-01-03, got %v", p.Expenses[0].Date)
	}
	if !p.Expenses[1].Date.Equal(d(2026, 1, 1)) {
		t.Errorf("second expense date: want 2026-01-01, got %v", p.Expenses[1].Date)
	}
}

func TestBuildReport_PersonRawIncomes(t *testing.T) {
	rawIncomes := []domain.Income{
		income("Alice", "Salary", 2000, d(2026, 1, 5), ""),
		income("Alice", "Bonus", 500, d(2026, 1, 1), ""),
	}
	persons := []domain.PersonData{
		{Person: domain.Person{Name: "Alice"}, Incomes: rawIncomes},
	}
	h := domain.HouseholdData{
		Incomes: rawIncomes,
	}

	r := presentation.BuildReport(h, persons)

	p := r.People[0]
	if len(p.Incomes) != 2 {
		t.Fatalf("raw incomes: want 2, got %d", len(p.Incomes))
	}
	// Date descending: Jan 5 first
	if !p.Incomes[0].Date.Equal(d(2026, 1, 5)) {
		t.Errorf("first income date: want 2026-01-05, got %v", p.Incomes[0].Date)
	}
}

func TestBuildReport_PersonCategoryGroupsPassedThrough(t *testing.T) {
	groups := []domain.CategoryGroup{
		{Name: "Fun", TargetPercent: 0.1, TargetAmount: 100, ActualSpent: 80, Delta: 20},
	}
	persons := []domain.PersonData{
		{
			Person:         domain.Person{Name: "Alice"},
			CategoryGroups: groups,
			Expenses:       []domain.Expense{expense("Alice", "Leisure", 80, d(2026, 1, 1), "")},
		},
	}
	h := domain.HouseholdData{
		Expenses: persons[0].Expenses,
	}

	r := presentation.BuildReport(h, persons)

	if len(r.People[0].CategoryGroups) != 1 {
		t.Fatalf("want 1 person category group, got %d", len(r.People[0].CategoryGroups))
	}
	if r.People[0].CategoryGroups[0].Name != "Fun" {
		t.Errorf("group name: want Fun, got %q", r.People[0].CategoryGroups[0].Name)
	}
}

// ── Empty inputs ──────────────────────────────────────────

func TestBuildReport_Empty(t *testing.T) {
	r := presentation.BuildReport(domain.HouseholdData{}, nil)

	if r.Household.Totals.Income.Value != 0 {
		t.Errorf("empty: income want 0, got %.2f", r.Household.Totals.Income.Value)
	}
	if len(r.People) != 0 {
		t.Errorf("empty: want 0 people, got %d", len(r.People))
	}
	if len(r.Household.ExpensesByCategory) != 0 {
		t.Errorf("empty: want 0 expense categories, got %d", len(r.Household.ExpensesByCategory))
	}
}

// ── Determinism ───────────────────────────────────────────

func TestBuildReport_Deterministic(t *testing.T) {
	// Running BuildReport twice on the same input should produce identical output.
	h := domain.HouseholdData{
		Expenses: []domain.Expense{
			expense("Zara", "Food", 100, d(2026, 1, 1), ""),
			expense("Alice", "Rent", 800, d(2026, 1, 1), ""),
			expense("Bob", "Transport", 50, d(2026, 1, 1), ""),
		},
		Incomes: []domain.Income{
			income("Zara", "Salary", 2000, d(2026, 1, 1), ""),
			income("Alice", "Salary", 3000, d(2026, 1, 1), ""),
		},
	}

	r1 := presentation.BuildReport(h, nil)
	r2 := presentation.BuildReport(h, nil)

	if len(r1.People) != len(r2.People) {
		t.Fatalf("people count differs between runs")
	}
	for i := range r1.People {
		if r1.People[i].Name != r2.People[i].Name {
			t.Errorf("people[%d] name differs: %q vs %q", i, r1.People[i].Name, r2.People[i].Name)
		}
	}

	if len(r1.Household.ExpensesByCategory) != len(r2.Household.ExpensesByCategory) {
		t.Fatal("ExpensesByCategory count differs between runs")
	}
	for i := range r1.Household.ExpensesByCategory {
		if r1.Household.ExpensesByCategory[i].Category != r2.Household.ExpensesByCategory[i].Category {
			t.Errorf("category[%d] differs between runs", i)
		}
	}
}
