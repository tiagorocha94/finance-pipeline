# household-finance-pipeline

A CLI pipeline that parses personal finance exports, aggregates them into a household view, and produces reports browsable via a local web dashboard.

## How it works

1. **Ingest** — reads raw files (`.csv` or `.xlsx`) from `data_dir/<month>/`, matching the configured people
2. **Parse** — each file represents one person; the filename (without extension) must match a name in `people`
3. **Aggregate** — merges all persons into a household view and optionally computes budget group targets against actual spending
4. **Export** — writes `output/<YYYY-MM>/data.json` and updates `output/manifest.json`

The dashboard (`output/index.html`) is a static SPA that reads those files. It lives in the repo and is never overwritten by the pipeline.

## Project structure

```
cmd/
  cli/main.go       ← pipeline runner
  server/main.go    ← local file server for the dashboard
output/
  index.html        ← dashboard SPA (tracked in repo)
  manifest.json     ← auto-maintained by pipeline (gitignored)
  YYYY-MM/
    data.json       ← generated per run (gitignored)
data/               ← input files, gitignored
config.json         ← local configuration, gitignored
```

## Configuration

Copy `config.example.json` to `config.json` and fill in your own values. The file is gitignored so your personal details stay local.

```json
{
  "data_dir": "./data",
  "output_dir": "output",
  "people": ["Alice", "Bob"],
  "category_groups": [
    {
      "name": "Essentials",
      "percent": "50%",
      "categories": ["Rent", "Groceries", "Transport", "Health"]
    },
    {
      "name": "Investments",
      "percent": "20%",
      "categories": ["Stocks", "Crypto"]
    },
    {
      "name": "Savings",
      "percent": "10%",
      "categories": ["Savings", "Emergency fund"]
    },
    {
      "name": "Fun",
      "percent": "10%",
      "categories": ["Restaurants", "Travel", "*"]
    }
  ]
}
```

| Field | Description |
|---|---|
| `data_dir` | Root directory where input files are read from |
| `output_dir` | Directory where `data.json` and `manifest.json` are written. Defaults to `output` |
| `people` | List of person names; input filenames must match these (without extension) |
| `category_groups` | Budget groups with a target % of total income and the expense categories that belong to each. Use `"*"` in a group's categories to catch any expense not matched by another group. Omit the field entirely to skip budget group calculations |

## Input file format

Place input files under `<data_dir>/<YYYY-MM>/`, e.g. `./data/2026-01/`:

- **CSV** — one row per transaction with columns: `type, amount, currency, category, date, note`
- **XLSX** — export from the finance app; expenses in the `Despesas` sheet, income in `Receita`

Each filename (e.g. `Alice.xlsx`, `Bob.csv`) must match a name in the `people` config field.

## Running the pipeline

```bash
# Current month (default)
make run

# Specific month
go run ./cmd/cli/main.go -month 2026-01
```

## Running tests

```bash
make test     # run all tests
make vet      # run go vet
```

## Viewing the dashboard

Start the file server (requires no additional installs — just Go):

```bash
go run ./cmd/server/main.go
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

Options:

```bash
go run ./cmd/server/main.go -port 9000          # custom port
go run ./cmd/server/main.go -dir ./other/path   # custom output directory
```

## Requirements

- Go 1.22+
- Internet access for the dashboard's Google Fonts (optional, degrades gracefully)