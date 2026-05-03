package pipeline

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"tiagorocha94/household-finance-pipeline/internal/aggregation"
	"tiagorocha94/household-finance-pipeline/internal/config"
	"tiagorocha94/household-finance-pipeline/internal/domain"
	"tiagorocha94/household-finance-pipeline/internal/ingestion"
	"tiagorocha94/household-finance-pipeline/internal/manifest"
	"tiagorocha94/household-finance-pipeline/internal/parsing"
	"tiagorocha94/household-finance-pipeline/internal/parsing/router"
	"tiagorocha94/household-finance-pipeline/internal/presentation"
)

// Pipeline orchestrates the full workflow:
// ingestion → parsing → aggregation → export
type Pipeline struct {
	cfg config.Config
	log *slog.Logger

	ingestor ingestion.Ingestor
	parser   parsing.Parser

	aggregator aggregation.Aggregator
	exporters  []presentation.Exporter
}

// New creates a fully configured pipeline.
// All dependencies are explicit (no registry, no init magic).
func New(
	cfg config.Config,
	ingestor ingestion.Ingestor,
	parsers []parsing.Parser,
	aggregator aggregation.Aggregator,
	exporters []presentation.Exporter,
	logger *slog.Logger,
) (*Pipeline, error) {

	if logger == nil {
		logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	if ingestor == nil {
		return nil, fmt.Errorf("ingestor is required")
	}
	if len(parsers) == 0 {
		return nil, fmt.Errorf("at least one parser is required")
	}
	if aggregator == nil {
		return nil, fmt.Errorf("aggregator is required")
	}
	if len(exporters) == 0 {
		return nil, fmt.Errorf("at least one exporter is required")
	}

	p := &Pipeline{
		cfg:        cfg,
		log:        logger,
		ingestor:   ingestor,
		parser:     router.New(parsers),
		aggregator: aggregator,
		exporters:  exporters,
	}

	p.log.Info("pipeline initialized")

	return p, nil
}

// Run executes the pipeline for a given month.
func (p *Pipeline) Run(month string) error {
	p.log.Info("pipeline started", "month", month)

	rawFiles, err := p.ingestor.Ingest(p.cfg, month)
	if err != nil {
		return err
	}

	p.log.Info("files ingested", "count", len(rawFiles))

	var parsed []domain.PersonData

	for _, file := range rawFiles {
		p.log.Debug("parsing file", "file", file.Name)

		data, err := p.parser.Parse(file)
		if err != nil {
			return fmt.Errorf("parse %s: %w", file.Name, err)
		}
		parsed = append(parsed, data)
	}

	p.log.Info("parsing completed", "persons", len(parsed))

	household, peopleData, err := p.aggregator.Aggregate(parsed)
	if err != nil {
		return err
	}

	p.log.Info("aggregation completed")

	report := buildReport(household, peopleData)

	for _, exporter := range p.exporters {
		name := fmt.Sprintf("%T", exporter)

		p.log.Info("exporting report", "exporter", name)

		outputs, err := exporter.Export(report)
		if err != nil {
			return err
		}

		if err := p.writeOutputs(month, outputs); err != nil {
			return err
		}
	}

	if err := manifest.Update("output"); err != nil {
		return fmt.Errorf("update manifest: %w", err)
	}

	p.log.Info("pipeline finished successfully")
	return nil
}

// buildReport converts aggregated data into reporting model.
// It accepts the full []PersonData so it can look up per-person CategoryGroups by name.
func buildReport(h domain.HouseholdData, persons []domain.PersonData) presentation.Report {
	householdExpense := map[string]float64{}
	householdIncome := map[string]float64{}

	personExpense := map[string]map[string]float64{}
	personIncome := map[string]map[string]float64{}

	var totalIncome, totalExpense float64

	// EXPENSES
	for _, e := range h.Expenses {
		totalExpense += e.Amount.Value

		householdExpense[e.Category] += e.Amount.Value

		if personExpense[e.PersonName] == nil {
			personExpense[e.PersonName] = map[string]float64{}
		}
		personExpense[e.PersonName][e.Category] += e.Amount.Value
	}

	// INCOMES
	for _, i := range h.Incomes {
		totalIncome += i.Amount.Value

		householdIncome[i.Source] += i.Amount.Value

		if personIncome[i.PersonName] == nil {
			personIncome[i.PersonName] = map[string]float64{}
		}
		personIncome[i.PersonName][i.Source] += i.Amount.Value
	}

	// HOUSEHOLD VIEW
	household := presentation.HouseholdView{
		Totals: presentation.Totals{
			Income:  domain.Money{Value: totalIncome, Currency: "EUR"},
			Expense: domain.Money{Value: totalExpense, Currency: "EUR"},
			Balance: domain.Money{Value: totalIncome - totalExpense, Currency: "EUR"},
		},
		CategoryGroups: h.CategoryGroups,
	}

	for k, v := range householdExpense {
		household.ExpensesByCategory = append(household.ExpensesByCategory, presentation.CategoryTotal{
			Category: k,
			Total:    domain.Money{Value: v, Currency: "EUR"},
		})
	}

	for k, v := range householdIncome {
		household.IncomeBySource = append(household.IncomeBySource, presentation.SourceTotal{
			Source: k,
			Total:  domain.Money{Value: v, Currency: "EUR"},
		})
	}

	// Build a name → PersonData lookup for category groups and raw transactions.
	personGroups := make(map[string][]domain.CategoryGroup, len(persons))
	personExpenses := make(map[string][]domain.Expense)
	personIncomes := make(map[string][]domain.Income)
	for _, pd := range persons {
		personGroups[pd.Person.Name] = pd.CategoryGroups
		personExpenses[pd.Person.Name] = pd.Expenses
		personIncomes[pd.Person.Name] = pd.Incomes
	}

	// PERSON VIEWS
	seen := map[string]bool{}
	var people []presentation.PersonView

	for name := range personExpense {
		seen[name] = true
		pv := buildPersonView(name, personExpense[name], personIncome[name], personGroups[name], personExpenses[name], personIncomes[name])
		people = append(people, pv)
	}
	for name := range personIncome {
		if seen[name] {
			continue
		}
		pv := buildPersonView(name, nil, personIncome[name], personGroups[name], personExpenses[name], personIncomes[name])
		people = append(people, pv)
	}

	return presentation.Report{
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
) presentation.PersonView {
	pv := presentation.PersonView{
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

	pv.Totals = presentation.Totals{
		Income:  domain.Money{Value: pIncome, Currency: "EUR"},
		Expense: domain.Money{Value: pExpense, Currency: "EUR"},
		Balance: domain.Money{Value: pIncome - pExpense, Currency: "EUR"},
	}

	for k, v := range expenses {
		pv.ExpensesByCategory = append(pv.ExpensesByCategory, presentation.CategoryTotal{
			Category: k,
			Total:    domain.Money{Value: v, Currency: "EUR"},
		})
	}

	for k, v := range incomes {
		pv.IncomeBySource = append(pv.IncomeBySource, presentation.SourceTotal{
			Source: k,
			Total:  domain.Money{Value: v, Currency: "EUR"},
		})
	}

	return pv
}

// writeOutputs handles final artifact writing.
func (p *Pipeline) writeOutputs(month string, outputs []presentation.Output) error {
	outputDir := filepath.Join("output", month)

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for _, o := range outputs {
		path := filepath.Join(outputDir, o.File.Name)

		if err := os.WriteFile(path, o.File.Content, 0644); err != nil {
			return fmt.Errorf("writing %s: %w", path, err)
		}

		p.log.Info("file written", "file", path, "bytes", len(o.File.Content))
	}

	return nil
}
