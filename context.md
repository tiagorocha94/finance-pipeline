# household-finance-pipeline — Repository Context

Quick reference for working in this codebase. Read this before making changes.

---

## Working on ideas.md items

When asked to implement an `ideas.md` item, follow the workflow defined in `AGENTS.md`:
implement → add/update tests → update `ideas.md` → update `context.md` → output all changed files → commit message.

---

## What it does

A Go CLI pipeline that reads personal finance exports (CSV or XLSX), aggregates them into a household view, and writes `data.json` per month. A standalone SPA (`output/index.html`) reads those files and renders an interactive dashboard served by a built-in Go file server.

---

## Repository layout

```
cmd/
  cli/main.go           entry point — parses -month flag, wires pipeline
  server/main.go        static file server for output/ (-dir, -port flags)

internal/
  config/
    config.go           Config struct, Load(), setDefaults(), validate()
    config_test.go
  domain/
    models.go           core value types: Money, Expense, Income, Person,
                        PersonData, HouseholdData, CategoryGroup
    types.go            RawFile, OutputFile
  ingestion/
    ingestion.go        Ingestor interface
    filesystem/
      filesystem.go     reads data/<month>/<Person>.<ext> files
  parsing/
    parsing.go          Parser interface
    csv/
      csv.go            CSV parser — columns: type,amount,currency,category,date,note
      csv_test.go
    xlsx/
      xlsx.go           XLSX parser — sheets: Despesas (expenses), Receita (income)
    router/
      router.go         dispatches to parser by file extension
  aggregation/
    aggregation.go      Aggregator interface
    household/
      household.go      Aggregate() + computeGroups() with catch-all "*" support
      household_test.go
  manifest/
    manifest.go         Update() — scans output dir, writes manifest.json
  pipeline/
    pipeline.go         New(), Run(), buildReport(), buildPersonView()
  presentation/
    model.go            Report, HouseholdView, PersonView, Totals, CategoryTotal,
                        SourceTotal — PersonView carries raw Expenses/Incomes slices
    build.go            BuildReport() + buildPersonView() — converts domain data
                        into the presentation model; all sorting lives here
    exporter.go         Exporter interface
    output.go           Output struct
    json/json.go        serialises Report → data.json
    markdown/markdown.go renders Report → report.md
    csv/csv.go          flat transactions.csv — one row per expense/income, no aggregation

output/
  index.html            SPA — tracked in repo, never overwritten by pipeline
  .gitignore            ignores everything in output/ except index.html and .gitignore

config.example.json     safe-to-commit template (generic names, no personal data)
ideas.md                30 prioritised improvement ideas (value × complexity)
```

---

## Configuration (`config.json`, gitignored)

| Field | Default | Description |
|---|---|---|
| `data_dir` | `"data"` | Root dir for input files |
| `output_dir` | `"output"` | Root dir for generated output |
| `people` | required | Person names; filenames must match |
| `category_groups` | optional | Budget groups — omit to skip calculations |

`category_groups` entry shape:
```json
{ "name": "Essentials", "percent": "50%", "categories": ["Rent", "Food", "*"] }
```
`"*"` is a catch-all for any expense not matched by another group. Only one group may use it. Validated on `config.Load()`.

---

## Data flow

```
filesystem.Ingestor.Ingest()
  └─ returns []RawFile (one per person, matched by name glob)

router.Router.Parse()          ← dispatches by extension
  ├─ csv.Parser.Parse()
  └─ xlsx.Parser.Parse()
  └─ returns []PersonData

household.Aggregator.Aggregate()
  ├─ merges into HouseholdData
  ├─ computes CategoryGroups per person and for the household
  └─ returns (HouseholdData, []PersonData, error)   ← enriched, not mutated input

presentation.BuildReport()    ← lives in internal/presentation/build.go
  ├─ builds presentation.Report from HouseholdData + []PersonData
  ├─ all slices sorted: people alphabetically, categories/sources by amount desc,
  │  raw transactions by date desc
  └─ PersonView.Expenses / .Incomes carry raw transactions for the SPA

json.Exporter.Export()         → output/<month>/data.json
markdown.Exporter.Export()     → output/<month>/report.md
csv.Exporter.Export()          → output/<month>/transactions.csv

manifest.Update(outputDir)     → output/manifest.json  (sorted newest-first)
```

---

## Input file formats

**CSV** (`data/<YYYY-MM>/<Person>.csv`):
```
type,amount,currency,category,date,note
expense,120.50,EUR,Food,2026-01-15,Supermarket
income,2500.00,EUR,Salary,2026-01-01,
```

**XLSX** (`data/<YYYY-MM>/<Person>.xlsx`):
- Sheet `Despesas` → expenses
- Sheet `Receita` → incomes
- Row 0: title (skipped), Row 1: headers (skipped), Row 2+: data
- Columns (0-based): 0=date, 1=category, 3=amount, 4=currency, 10=note
- Category names are normalised: first letter uppercased (`"seguros"` → `"Seguros"`)
- Supported date formats: `2006-01-02 15:04:05`, `2006-01-02`, `1/2/2006`, `01/02/2006`, `2/1/2006`, `02/01/2006`, `01-02-06`

---

## Output

Each pipeline run produces:
- `output/<YYYY-MM>/data.json` — full `presentation.Report` as JSON
- `output/<YYYY-MM>/report.md` — markdown summary
- `output/<YYYY-MM>/transactions.csv` — flat expense+income rows, no aggregation
- `output/manifest.json` — `{ "months": ["2026-02", "2026-01", ...] }` newest-first

The SPA reads `manifest.json` on boot, then fetches each month's `data.json` on demand. All data is cached in memory after first fetch.

---

## SPA (`output/index.html`)

Two pages, vanilla JS, no build step, no dependencies beyond Google Fonts:

**Overview** (default landing page):
- 4 KPI cards: total income/expense/balance, avg savings rate
- Monthly averages bar chart
- Over-time SVG chart: grouped bars (income green, expense red) + monthly balance line (blue, solid) + cumulative balance line (gold, dotted) with Y-axis labels and hover tooltip
- Budget groups over-time table: months × groups, avg row at bottom
- Cumulative pie charts: expenses by category, income by source
- Per-person cards: totals, monthly averages bar chart, mini over-time chart, pies

**Monthly**:
- 3 KPI cards + summary bar chart
- Household breakdown section with Table/Charts toggle; Charts shows pie charts
- If `CategoryGroups` present: tab toggle Categories / Budget Groups in both Table and Charts views
- People breakdown grid (2 columns, 50% each)
- Transactions section: per-person sortable/filterable tables for expenses and incomes

**Key JS functions:**
- `navigate(page)` — switches page, shows/hides month list in sidebar
- `loadMonth(month)` — fetches and caches `data.json`, calls `renderReport()`
- `renderOverview()` — loads all months in parallel, aggregates, renders
- `renderSection(scope, data, totalIn, totalEx, compact)` — shared renderer for household and person breakdowns
- `renderPersonTransactions(p, pi)` — transaction tables with sort/filter state in `txSortState`
- `switchView(btn, showId, hideId)` — Table/Charts toggle (uses direct getElementById)
- `switchTab(btn, targetId)` — Categories/Budget Groups toggle (scoped to bar's parentElement)
- `fmoney(v)` — Portuguese locale: `1.238,93 €`
- `fmoneyShort(v)` — compact axis labels: `1.2k`

---

## Running

```bash
make run                              # pipeline, current month
go run ./cmd/cli/main.go -month 2026-01  # specific month
make serve                            # dashboard at http://localhost:8080
go run ./cmd/server/main.go -port 9000 -dir ./output
make test                             # all tests
make vet                              # go vet
```

---

## Tests

| Package | File | Coverage |
|---|---|---|
| `internal/config` | `config_test.go` | Load, defaults, all validation branches |
| `internal/parsing/csv` | `csv_test.go` | Happy path, person name, all error cases |
| `internal/aggregation/household` | `household_test.go` | Aggregate, computeGroups, catch-all, per-person, mutation guard |

No tests yet for: XLSX parser, markdown exporter, manifest, router. End-to-end pipeline test pending. See `ideas.md` items 1, 2.

---

## CI / Linting

`.github/workflows/ci.yml` — runs on push/PR to `main`:
1. `go mod verify`
2. `go vet ./...`
3. `go test -race -count=1 ./...`
4. `golangci-lint` (v2.12.1) with: errcheck, govet, ineffassign, staticcheck, unused

Suppression pattern in code: `//nolint: errcheck` used on `f.Close()` in xlsx parser and temp file close in config test.

---

## Known design decisions & gotchas

- **Currency hardcoded to `"EUR"`** in `buildReport` — mixed-currency files will silently produce wrong totals (ideas.md #14)
- **XLSX category normalisation** only title-cases the first character — full case-insensitive matching is not implemented
- **`manifest.Update`** will error if `output/` doesn't exist on first run — `os.MkdirAll` is called in `writeOutputs` before it, so this is safe as long as at least one exporter runs first
- **SPA over-time chart** recomputes the Y-axis scale after bars are already drawn; the first pass uses per-month values only, then extends to include cumulative balances for the line overlay — bar positions use `yPx`, line positions use `yPx2`
- **Transaction data** is embedded in the HTML as a hidden `<script type="application/json">` element per table so sort/filter can re-render without re-fetching JSON
- **People order** in the report is alphabetical (sorted in `buildReport`); category order is by amount descending; raw transactions are date descending