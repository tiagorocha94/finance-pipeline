package household_test

import (
	"testing"

	"tiagorocha94/household-finance-pipeline/internal/aggregation/household"
	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/domain"
)

// helpers

func expense(category string, amount float64) domain.Expense {
	return domain.Expense{
		PersonName: "Alice",
		Category:   category,
		Amount:     domain.Money{Value: amount, Currency: "EUR"},
	}
}

func income(amount float64) domain.Income {
	return domain.Income{
		PersonName: "Alice",
		Source:     "Salary",
		Amount:     domain.Money{Value: amount, Currency: "EUR"},
	}
}

func cfg(groups ...config.CategoryGroup) config.Config {
	return config.Config{
		DataDir:        "./data",
		People:         []string{"Alice"},
		CategoryGroups: groups,
	}
}

func group(name, pct string, cats ...string) config.CategoryGroup {
	return config.CategoryGroup{Name: name, Percent: pct, Categories: cats}
}

// ── Aggregate tests ───────────────────────────────────────

func TestAggregate_NoCategories(t *testing.T) {
	a := household.New(cfg())
	data := []domain.PersonData{
		{
			Person:   domain.Person{Name: "Alice"},
			Expenses: []domain.Expense{expense("Food", 100)},
			Incomes:  []domain.Income{income(500)},
		},
	}

	hh, enriched, err := a.Aggregate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(hh.CategoryGroups) != 0 {
		t.Errorf("expected no category groups, got %d", len(hh.CategoryGroups))
	}
	if len(enriched[0].CategoryGroups) != 0 {
		t.Errorf("expected no person category groups, got %d", len(enriched[0].CategoryGroups))
	}
}

func TestAggregate_BasicGroups(t *testing.T) {
	c := cfg(
		group("Essentials", "50%", "Food", "Rent"),
		group("Fun", "20%", "Leisure"),
	)
	a := household.New(c)

	data := []domain.PersonData{
		{
			Person: domain.Person{Name: "Alice"},
			Expenses: []domain.Expense{
				expense("Food", 200),
				expense("Rent", 500),
				expense("Leisure", 100),
			},
			Incomes: []domain.Income{income(2000)},
		},
	}

	hh, _, err := a.Aggregate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	groupMap := make(map[string]domain.CategoryGroup)
	for _, g := range hh.CategoryGroups {
		groupMap[g.Name] = g
	}

	ess := groupMap["Essentials"]
	if ess.ActualSpent != 700 {
		t.Errorf("Essentials: want 700, got %.2f", ess.ActualSpent)
	}
	if ess.TargetAmount != 1000 { // 50% of 2000
		t.Errorf("Essentials target: want 1000, got %.2f", ess.TargetAmount)
	}
	if ess.Delta != 300 { // under budget
		t.Errorf("Essentials delta: want 300, got %.2f", ess.Delta)
	}

	fun := groupMap["Fun"]
	if fun.ActualSpent != 100 {
		t.Errorf("Fun: want 100, got %.2f", fun.ActualSpent)
	}
}

func TestAggregate_CatchAll(t *testing.T) {
	c := cfg(
		group("Essentials", "50%", "Food"),
		group("Others", "20%", "*"),
	)
	a := household.New(c)

	data := []domain.PersonData{
		{
			Person: domain.Person{Name: "Alice"},
			Expenses: []domain.Expense{
				expense("Food", 300),
				expense("Unknown1", 50),
				expense("Unknown2", 75),
			},
			Incomes: []domain.Income{income(1000)},
		},
	}

	hh, _, err := a.Aggregate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	groupMap := make(map[string]domain.CategoryGroup)
	for _, g := range hh.CategoryGroups {
		groupMap[g.Name] = g
	}

	if groupMap["Essentials"].ActualSpent != 300 {
		t.Errorf("Essentials: want 300, got %.2f", groupMap["Essentials"].ActualSpent)
	}
	// Unknown1 + Unknown2 should fall into catch-all
	if groupMap["Others"].ActualSpent != 125 {
		t.Errorf("Others (catch-all): want 125, got %.2f", groupMap["Others"].ActualSpent)
	}
}

func TestAggregate_CatchAllDoesNotOverrideExplicit(t *testing.T) {
	// An expense that is explicitly mapped should NOT also go to the catch-all.
	c := cfg(
		group("Essentials", "50%", "Food"),
		group("Others", "30%", "*"),
	)
	a := household.New(c)

	data := []domain.PersonData{
		{
			Person: domain.Person{Name: "Alice"},
			Expenses: []domain.Expense{
				expense("Food", 400),
			},
			Incomes: []domain.Income{income(1000)},
		},
	}

	hh, _, err := a.Aggregate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, g := range hh.CategoryGroups {
		if g.Name == "Others" && g.ActualSpent != 0 {
			t.Errorf("Others (catch-all) should be 0 when all categories are mapped, got %.2f", g.ActualSpent)
		}
	}
}

func TestAggregate_PerPersonGroups(t *testing.T) {
	c := cfg(group("Savings", "10%", "Bank"))
	a := household.New(c)

	data := []domain.PersonData{
		{
			Person:   domain.Person{Name: "Alice"},
			Expenses: []domain.Expense{expense("Bank", 100)},
			Incomes:  []domain.Income{income(1000)},
		},
		{
			Person:   domain.Person{Name: "Bob"},
			Expenses: []domain.Expense{expense("Bank", 50)},
			Incomes:  []domain.Income{income(500)},
		},
	}

	_, enriched, err := a.Aggregate(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, p := range enriched {
		if len(p.CategoryGroups) == 0 {
			t.Errorf("person %s: expected category groups", p.Person.Name)
		}
		g := p.CategoryGroups[0]
		switch p.Person.Name {
		case "Alice":
			if g.TargetAmount != 100 { // 10% of 1000
				t.Errorf("Alice target: want 100, got %.2f", g.TargetAmount)
			}
		case "Bob":
			if g.TargetAmount != 50 { // 10% of 500
				t.Errorf("Bob target: want 50, got %.2f", g.TargetAmount)
			}
		}
	}
}

func TestAggregate_EnrichedSliceIsNotMutatedInput(t *testing.T) {
	// Confirm the original input slice is not mutated (value-copy semantics).
	c := cfg(group("G", "10%", "X"))
	a := household.New(c)

	original := []domain.PersonData{
		{
			Person:   domain.Person{Name: "Alice"},
			Expenses: []domain.Expense{expense("X", 50)},
			Incomes:  []domain.Income{income(500)},
		},
	}

	_, _, err := a.Aggregate(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(original[0].CategoryGroups) != 0 {
		t.Error("original input was mutated: CategoryGroups should remain nil")
	}
}
