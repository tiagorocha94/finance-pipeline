package markdown

import (
	"bytes"
	"fmt"
	"math"
	"strings"
	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
	"time"
)

// Exporter generates Markdown output.
type Exporter struct{}

func New() *Exporter {
	return &Exporter{}
}

func (e *Exporter) Export(r presentation.Report) ([]presentation.Output, error) {
	var buf bytes.Buffer
	w := func(s string) { buf.WriteString(s) }
	wf := func(format string, args ...any) { buf.WriteString(fmt.Sprintf(format, args...)) }
	nl := func() { buf.WriteByte('\n') }

	income := r.Household.Totals.Income.Value
	expense := r.Household.Totals.Expense.Value
	balance := r.Household.Totals.Balance.Value
	savingsRate := pct(balance, income)
	coverageRatio := 0.0
	if expense > 0 {
		coverageRatio = income / expense
	}

	// ── Title block ──────────────────────────────────────────────
	w("# 🏠 Household Financial Report\n\n")
	wf("> Generated on **%s**\n\n", time.Now().Format("2 January 2006"))
	w("---\n\n")

	// ── Household summary ────────────────────────────────────────
	w("## Summary\n\n")
	w("| Metric | Value |\n")
	w("|--------|-------|\n")
	wf("| 💰 Total Income     | `%s` |\n", fmtMoney(income))
	wf("| 💸 Total Expenses   | `%s` |\n", fmtMoney(expense))
	wf("| 📊 Net Balance      | `%s` |\n", fmtMoney(balance))
	wf("| 🐖 Savings Rate     | `%.1f%%` |\n", savingsRate)
	wf("| 🛡️  Income Coverage  | `%.2fx` expenses |\n", coverageRatio)
	wf("| 👥 People Tracked   | `%d` |\n", len(r.People))
	nl()

	// Savings rate narrative
	switch {
	case savingsRate >= 30:
		wf("> ✅ **Healthy savings rate** of %.1f%% — well above the recommended 20%%.\n\n", savingsRate)
	case savingsRate >= 20:
		wf("> ✅ **Good savings rate** of %.1f%% — at the recommended threshold.\n\n", savingsRate)
	case savingsRate >= 10:
		wf("> ⚠️  **Moderate savings rate** of %.1f%% — consider reviewing discretionary spending.\n\n", savingsRate)
	case savingsRate >= 0:
		wf("> ⚠️  **Low savings rate** of %.1f%% — income barely covers expenses.\n\n", savingsRate)
	default:
		wf("> ❌ **Negative balance** of %s — expenses exceed income.\n\n", fmtMoney(balance))
	}

	w("---\n\n")

	// ── Expenses by category ─────────────────────────────────────
	w("## Expenses by Category\n\n")
	if len(r.Household.ExpensesByCategory) == 0 {
		w("_No expense data recorded._\n\n")
	} else {
		topCat, topVal := findTopCategory(r.Household.ExpensesByCategory)
		wf("> 🔺 Largest expense: **%s** at `%s` (%.1f%% of spending)\n\n",
			topCat, fmtMoney(topVal), pct(topVal, expense))

		w("| Category | Amount | Share | Distribution |\n")
		w("|----------|-------:|------:|--------------|\n")
		for _, c := range r.Household.ExpensesByCategory {
			share := pct(c.Total.Value, expense)
			bar := sparkBar(share, 20)
			wf("| %s | `%s` | %.1f%% | %s |\n", c.Category, fmtMoney(c.Total.Value), share, bar)
		}
		nl()
	}

	// ── Income by source ─────────────────────────────────────────
	w("## Income by Source\n\n")
	if len(r.Household.IncomeBySource) == 0 {
		w("_No income data recorded._\n\n")
	} else {
		w("| Source | Amount | Share | Distribution |\n")
		w("|--------|-------:|------:|--------------|\n")
		for _, s := range r.Household.IncomeBySource {
			share := pct(s.Total.Value, income)
			bar := sparkBar(share, 20)
			wf("| %s | `%s` | %.1f%% | %s |\n", s.Source, fmtMoney(s.Total.Value), share, bar)
		}
		nl()
	}

	w("---\n\n")

	// ── People comparison ─────────────────────────────────────────
	if len(r.People) > 0 {
		w("## People at a Glance\n\n")

		// Header row — dynamic number of people
		headerCols := "| Metric |"
		sepCols := "|--------|"
		for _, p := range r.People {
			headerCols += fmt.Sprintf(" %s |", p.Name)
			sepCols += "-------:|"
		}
		w(headerCols + "\n")
		w(sepCols + "\n")

		// Income row
		row := "| 💰 Income |"
		for _, p := range r.People {
			row += fmt.Sprintf(" `%s` |", fmtMoney(p.Totals.Income.Value))
		}
		w(row + "\n")

		// Expenses row
		row = "| 💸 Expenses |"
		for _, p := range r.People {
			row += fmt.Sprintf(" `%s` |", fmtMoney(p.Totals.Expense.Value))
		}
		w(row + "\n")

		// Balance row
		row = "| 📊 Balance |"
		for _, p := range r.People {
			row += fmt.Sprintf(" `%s` |", fmtMoney(p.Totals.Balance.Value))
		}
		w(row + "\n")

		// Savings rate row
		row = "| 🐖 Savings Rate |"
		for _, p := range r.People {
			row += fmt.Sprintf(" %.1f%% |", pct(p.Totals.Balance.Value, p.Totals.Income.Value))
		}
		w(row + "\n")

		// Income share of household row
		row = "| 🏠 Income Share |"
		for _, p := range r.People {
			row += fmt.Sprintf(" %.1f%% |", pct(p.Totals.Income.Value, income))
		}
		w(row + "\n")

		// Expense share of household row
		row = "| 📉 Expense Share |"
		for _, p := range r.People {
			row += fmt.Sprintf(" %.1f%% |", pct(p.Totals.Expense.Value, expense))
		}
		w(row + "\n")
		nl()

		w("---\n\n")

		// ── Per-person detail ────────────────────────────────────
		w("## Per Person Breakdown\n\n")

		for _, p := range r.People {
			pIncome := p.Totals.Income.Value
			pExpense := p.Totals.Expense.Value
			pBalance := p.Totals.Balance.Value
			pSavings := pct(pBalance, pIncome)
			hContrib := pct(pIncome, income)

			wf("### %s\n\n", p.Name)

			// Mini summary
			w("| | |\n")
			w("|-|-|\n")
			wf("| Income | `%s` |\n", fmtMoney(pIncome))
			wf("| Expenses | `%s` |\n", fmtMoney(pExpense))
			wf("| Balance | `%s` |\n", fmtMoney(pBalance))
			wf("| Savings Rate | `%.1f%%` |\n", pSavings)
			wf("| Household Income Contribution | `%.1f%%` |\n", hContrib)
			nl()

			// Expenses table
			if len(p.ExpensesByCategory) > 0 {
				topCat, topVal := findTopCategory(p.ExpensesByCategory)
				wf("**Expenses** — top spend: **%s** (`%s`)\n\n", topCat, fmtMoney(topVal))
				w("| Category | Amount | of Own Expenses | of Household |\n")
				w("|----------|-------:|----------------:|-------------:|\n")
				for _, c := range p.ExpensesByCategory {
					ownShare := pct(c.Total.Value, pExpense)
					hhShare := pct(c.Total.Value, expense)
					bar := sparkBar(ownShare, 15)
					wf("| %s | `%s` | %s %.1f%% | %.1f%% |\n",
						c.Category, fmtMoney(c.Total.Value), bar, ownShare, hhShare)
				}
				nl()
			} else {
				w("_No expenses recorded._\n\n")
			}

			// Income table
			if len(p.IncomeBySource) > 0 {
				w("**Income**\n\n")
				w("| Source | Amount | of Own Income | of Household |\n")
				w("|--------|-------:|--------------:|-------------:|\n")
				for _, s := range p.IncomeBySource {
					ownShare := pct(s.Total.Value, pIncome)
					hhShare := pct(s.Total.Value, income)
					bar := sparkBar(ownShare, 15)
					wf("| %s | `%s` | %s %.1f%% | %.1f%% |\n",
						s.Source, fmtMoney(s.Total.Value), bar, ownShare, hhShare)
				}
				nl()
			} else {
				w("_No income recorded._\n\n")
			}

			w("---\n\n")
		}
	}

	// ── Footer ───────────────────────────────────────────────────
	w("_Report generated automatically. All monetary values in the household's base currency._\n")

	return []presentation.Output{
		{
			Name: "markdown",
			File: domain.OutputFile{
				Name:    "report.md",
				Content: buf.Bytes(),
			},
		},
	}, nil
}

// fmtMoney formats a float as a right-aligned money string.
// Negative values use a proper minus sign.
func fmtMoney(v float64) string {
	if v < 0 {
		return fmt.Sprintf("−%.2f", -v)
	}
	return fmt.Sprintf("%.2f", v)
}

// pct returns what percentage a is of b, returning 0 if b is zero.
func pct(a, b float64) float64 {
	if b == 0 {
		return 0
	}
	return (a / b) * 100
}

// sparkBar renders a compact ASCII bar proportional to share (0–100).
// width is the maximum number of filled characters.
func sparkBar(share float64, width int) string {
	const filled = '█'
	const empty = '░'
	n := int(math.Round(share / 100 * float64(width)))
	if n > width {
		n = width
	}
	return strings.Repeat(string(filled), n) + strings.Repeat(string(empty), width-n)
}

// findTopCategory returns the name and value of the CategoryTotal with the highest value.
func findTopCategory(cats []presentation.CategoryTotal) (string, float64) {
	if len(cats) == 0 {
		return "", 0
	}
	top := cats[0]
	for _, c := range cats[1:] {
		if c.Total.Value > top.Total.Value {
			top = c
		}
	}
	return top.Category, top.Total.Value
}
