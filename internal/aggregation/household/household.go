package household

import (
	"fmt"

	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/domain"
)

type Aggregator struct {
	cfg config.Config
}

func New(cfg config.Config) *Aggregator {
	return &Aggregator{cfg: cfg}
}

// Aggregate merges all PersonData into a HouseholdData and computes category
// groups (when configured). It returns the enriched PersonData slice so callers
// have access to per-person CategoryGroups without relying on in-place mutation.
func (a *Aggregator) Aggregate(data []domain.PersonData) (domain.HouseholdData, []domain.PersonData, error) {
	enriched := make([]domain.PersonData, len(data))
	copy(enriched, data)

	var persons []domain.Person
	var allExpenses []domain.Expense
	var allIncomes []domain.Income

	for i, d := range enriched {
		persons = append(persons, d.Person)
		allExpenses = append(allExpenses, d.Expenses...)
		allIncomes = append(allIncomes, d.Incomes...)

		if len(a.cfg.CategoryGroups) == 0 {
			continue
		}

		groups, err := computeGroups(a.cfg, d.Expenses, d.Incomes)
		if err != nil {
			return domain.HouseholdData{}, nil, fmt.Errorf("category groups for %q: %w", d.Person.Name, err)
		}
		enriched[i].CategoryGroups = groups
	}

	household := domain.HouseholdData{
		Persons:  persons,
		Expenses: allExpenses,
		Incomes:  allIncomes,
	}

	if len(a.cfg.CategoryGroups) > 0 {
		groups, err := computeGroups(a.cfg, allExpenses, allIncomes)
		if err != nil {
			return domain.HouseholdData{}, nil, fmt.Errorf("household category groups: %w", err)
		}
		household.CategoryGroups = groups
	}

	return household, enriched, nil
}

// computeGroups builds a []CategoryGroup from the given expenses and incomes,
// using the category group definitions in cfg.
//
// A group whose Categories slice contains "*" is treated as a catch-all: any
// expense whose category is not explicitly mapped to another group is counted
// towards it. Only one catch-all group is supported; if multiple are defined
// the first one wins.
func computeGroups(cfg config.Config, expenses []domain.Expense, incomes []domain.Income) ([]domain.CategoryGroup, error) {
	totalIncome := sumIncomes(incomes)

	// Build a lookup: expense category name → budget group name.
	// Also record the catch-all group name, if any.
	expenseToBudgetGroup := make(map[string]string)
	catchAllGroup := ""
	for _, g := range cfg.CategoryGroups {
		for _, cat := range g.Categories {
			if cat == "*" {
				if catchAllGroup == "" {
					catchAllGroup = g.Name
				}
			} else {
				expenseToBudgetGroup[cat] = g.Name
			}
		}
	}

	// Accumulate actual spending per budget group.
	// Expenses with no explicit mapping fall into the catch-all group (if set).
	actualSpent := make(map[string]float64, len(cfg.CategoryGroups))
	for _, e := range expenses {
		if group, ok := expenseToBudgetGroup[e.Category]; ok {
			actualSpent[group] += e.Amount.Value
		} else if catchAllGroup != "" {
			actualSpent[catchAllGroup] += e.Amount.Value
		}
	}

	groups := make([]domain.CategoryGroup, 0, len(cfg.CategoryGroups))
	for _, g := range cfg.CategoryGroups {
		pct, err := g.ParsedPercent()
		if err != nil {
			return nil, err
		}
		target := totalIncome * pct
		actual := actualSpent[g.Name]
		groups = append(groups, domain.CategoryGroup{
			Name:          g.Name,
			TargetPercent: pct,
			TargetAmount:  target,
			ActualSpent:   actual,
			Delta:         target - actual,
		})
	}

	return groups, nil
}

func sumIncomes(incomes []domain.Income) float64 {
	var total float64
	for _, i := range incomes {
		total += i.Amount.Value
	}
	return total
}
