# Ideas & Improvements

Items are scored on **value** (impact on correctness, usability, or maintainability) and **complexity** (effort to implement). High-value / low-complexity items should be tackled first.

When implementing an item follow the workflow in `AGENTS.md`: implement → tests → update `ideas.md` → update `context.md` → output all changed files → commit message.

| # | Name | Type | Description | Value | Complexity |
|---|------|------|-------------|-------|------------|
| 1 | Unit tests — XLSX parser | Testing | The XLSX parser has no test coverage. Table-driven tests with fixture files would catch regressions when date formats or column layouts change. | High | Medium |
| 2 | Integration test — pipeline | Testing | A single end-to-end test that runs the pipeline against fixture CSV/XLSX files and asserts the output JSON matches a golden file. Catches wiring regressions. | High | Medium |
| 3 | `slog` levels via flag | Enhancement | Log level is hardcoded to `slog.LevelError` in `main.go`. Add a `-v` / `-log-level` flag so debug output can be enabled without editing source. | Medium | Low |
| 4 | Percent validation warning | Enhancement | If configured `category_groups` percents do not sum to 100%, log a warning. Useful to catch typos (e.g. forgetting a group). | Medium | Low |
| 5 | Create output dir on first run | Enhancement | `manifest.Update` errors if `output/` doesn't exist. Add `os.MkdirAll` before the scan so the first run doesn't require the directory to exist manually. | Medium | Low |
| 6 | Multi-currency support | Enhancement | `BuildReport` hardcodes `Currency: "EUR"`. Mixed-currency files silently produce wrong totals. Either enforce a single base currency via config or add conversion (requires exchange-rate source). | High | High |
| 7 | OFX / QIF parser | Enhancement | Support OFX (Open Financial Exchange) and QIF files — widely exported by banks and desktop apps. Same `Parser` interface, new package. | Medium | Medium |
| 8 | Google Sheets ingestor | Enhancement | An ingestor that reads directly from a Google Sheet (via Sheets API) instead of local files. Would remove the manual export step entirely. | High | High |
| 9 | Markdown exporter — transactions | Enhancement | The Markdown report omits raw transactions. Add an optional per-person transactions section, controlled by a config flag. | Low | Low |
| 10 | SPA — month URL routing | Enhancement | Deep-link to a specific month via URL hash (`#2026-01`) so sharing or bookmarking a monthly report is possible. | Medium | Medium |
| 11 | SPA — keyboard navigation | Enhancement | Allow switching months with arrow keys when the sidebar is focused. Small UX improvement for keyboard users. | Low | Low |
| 12 | SPA — print / PDF export | Enhancement | A print stylesheet that hides the sidebar and renders the monthly report cleanly, enabling browser-native PDF export. | Medium | Low |
| 13 | Docker / dev container | DevEx | A `Dockerfile` and/or `.devcontainer` so contributors can get a working environment without installing Go locally. | Low | Medium |
| 14 | Person-level data dir override | Enhancement | Allow a person's files to live in a different directory (e.g. shared network drive vs local). Edge case but relevant for multi-device households. | Low | High |
| 15 | Incremental re-run | Enhancement | Skip re-parsing months whose `data.json` is newer than the input files. Useful once there are many months of data. | Low | High |
| 16 | Budget group carry-over | Enhancement | Optionally roll unspent budget from one month into the next for groups like Savings. Requires state across months — significant design work. | Medium | High |