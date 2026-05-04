# Ideas & Improvements

Items are tagged by type and scored on **value** (impact on correctness, usability, or maintainability) and **complexity** (effort to implement). High-value / low-complexity items should be tackled first.

| # | Name | Type | Description | Value | Complexity |
|---|------|------|-------------|-------|------------|
| 1 | ~~Deterministic report order~~ | Bug | ✅ Done. All slices sorted: people alphabetically, categories/sources by amount desc, transactions by date desc. | High | Low |
| 2 | `OutputDir` in config | Architecture | `"output"` is hardcoded in two places (`pipeline.go` and `manifest.Update` call). Add `output_dir` to `Config` and thread it through. | High | Low |
| 3 | ~~Move `buildReport` out of pipeline~~ | Architecture | ✅ Done. Moved to `internal/presentation/build.go` as exported `BuildReport()`. | Medium | Low |
| 4 | Fix `xlsx` bytes→reader | Bug | `excelize.OpenReader(strings.NewReader(string(file.Content)))` does an unnecessary `[]byte → string → Reader` round-trip. Use `bytes.NewReader(file.Content)` directly. | Low | Low |
| 5 | Remove unused `minCols` constant | Cleanup | `minCols` in `xlsx.go` is defined but never referenced. | Low | Low |
| 6 | Config validation | Enhancement | Validate config on load: non-empty `people`, valid `data_dir`, percents that sum ≤ 100 (warn, not error), no duplicate group names. Return clear errors before the pipeline starts. | High | Low |
| 7 | `cmd/server` flags | Enhancement | The actual repo's server lost `-port` and `-dir` flags compared to the spec. Restore them so the server is configurable without editing source. | Medium | Low |
| 8 | Unit tests — parsers | Testing | CSV and XLSX parsers have zero test coverage. Table-driven tests with fixture files would catch regressions when date formats or column layouts change. | High | Medium |
| 9 | Unit tests — aggregator | Testing | `computeGroups` (including catch-all `*` logic) has no tests. Fast pure-function tests, no I/O needed. | High | Low |
| 10 | Unit tests — buildReport | Testing | `buildReport` / `buildPersonView` have no tests. Determinism fix (#1) should come first. | High | Medium |
| 11 | Integration test — pipeline | Testing | A single end-to-end test that runs the pipeline against fixture CSV/XLSX files and asserts the output JSON matches a golden file. Catches wiring regressions. | High | Medium |
| 12 | `slog` levels via flag | Enhancement | Log level is hardcoded to `slog.LevelError` in `main.go`. Add a `-v` / `-log-level` flag so debug output can be enabled without editing source. | Medium | Low |
| 13 | Stale router comment | Cleanup | `router.go` has a comment "We assume parser is CSV/JSON-specific and inferred externally via type…" that no longer reflects reality. | Low | Low |
| 14 | Multi-currency support | Enhancement | `buildReport` hardcodes `Currency: "EUR"`. If expenses and incomes carry mixed currencies the totals are silently wrong. Either enforce a single base currency or add conversion (requires exchange-rate source). | High | High |
| 15 | OFX / QIF parser | Enhancement | Support OFX (Open Financial Exchange) and QIF files — widely exported by banks and apps like Quicken / iBank. Same `Parser` interface, new package. | Medium | Medium |
| 16 | Google Sheets ingestor | Enhancement | An ingestor that reads directly from a Google Sheet (via Sheets API) instead of local files. Would remove the manual export step. | High | High |
| 17 | Markdown exporter — transactions | Enhancement | The Markdown report currently omits raw transactions. Add an optional per-person transactions section, controlled by a config flag. | Low | Low |
| 18 | CSV exporter | Enhancement | Export a flat `transactions.csv` alongside `data.json` for users who want to open data in Excel / Numbers. | Medium | Low |
| 19 | Category aliases | Enhancement | Allow multiple spellings of the same category to map to one canonical name (e.g. `"seguros"` and `"Seguros"` → `"Seguros"`). Currently case-sensitive matching silently drops mismatches. | High | Low |
| 20 | Percent validation warning | Enhancement | If configured `category_groups` percents do not sum to 100%, log a warning. Useful to catch typos (e.g. forgetting a group). | Medium | Low |
| 21 | `output_dir` gitignore helper | Enhancement | `manifest.Update` should create the output dir if it doesn't exist instead of failing silently when run for the first time. | Medium | Low |
| 22 | SPA — month URL routing | Enhancement | Deep-link to a specific month via URL hash (`#2026-01`) so sharing or bookmarking a monthly report is possible. | Medium | Medium |
| 23 | SPA — keyboard navigation | Enhancement | Allow switching months with arrow keys when the sidebar is focused. Small UX improvement for keyboard users. | Low | Low |
| 24 | SPA — print / PDF export | Enhancement | A print stylesheet that hides the sidebar and renders the monthly report cleanly, enabling browser-native PDF export. | Medium | Low |
| 25 | Docker / dev container | DevEx | A `Dockerfile` and/or `.devcontainer` so contributors can get a working environment without installing Go locally. | Low | Medium |
| 26 | GitHub Actions CI | DevEx | A workflow that runs `go vet`, `staticcheck`, and tests on push. Keeps the repo green automatically. | Medium | Low |
| 27 | `golangci-lint` config | DevEx | Add a `.golangci.yml` with a minimal linter set (errcheck, unused, govet, staticcheck). Catches issues like the unused `minCols` constant automatically. | Medium | Low |
| 28 | Person-level `output_dir` override | Enhancement | Allow a person's files to live in a different directory (e.g. shared network drive vs local). Edge case, but relevant for multi-device households. | Low | High |
| 29 | Incremental re-run | Enhancement | Skip re-parsing months whose `data.json` is newer than the input files. Useful once there are many months. | Low | High |
| 30 | Budget group carry-over | Enhancement | Optionally roll unspent budget from one month into the next for groups like Savings. Requires state across months — significant design work. | Medium | High |