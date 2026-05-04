# AGENTS.md

Instructions for AI assistants working on this repository.
Read this file at the start of every session alongside `context.md`.

---

## ideas.md workflow

When asked to implement an item from `ideas.md` (e.g. "#5", "do idea 12"):

1. **Implement** the work described in the item, following the architecture and conventions in `context.md`
2. **Tests** — add or update unit tests covering the new or changed behaviour; if existing tests are affected, fix them
3. **Update `ideas.md`** — remove the completed item entirely and renumber the remaining items
4. **Update `context.md`** — reflect any structural or behavioural changes (new files, changed data flow, new gotchas, etc.)
5. **Output all changed files** in full — every file that was created or modified, including `ideas.md` and `context.md`
6. **Provide a commit message** — conventional format, referencing the idea number

Do not skip steps 3–6 even for small changes.

---

## General conventions

- Module path: `tiagorocha94/household-finance-pipeline`
- Go version: see `go.mod`
- All output slices in `BuildReport` are sorted (people alphabetically, amounts descending, transactions by date descending) — do not break this
- Currency is hardcoded to `"EUR"` throughout `BuildReport` — do not silently change this
- Config is validated on `Load()` — new config fields should have defaults set in `setDefaults()` and validated in `validate()` or `validateCategoryGroups()`
- Tests use the standard library `testing` package only — no third-party assertion libraries
- Linter: `golangci-lint` with the config in `.golangci.yml`; use `//nolint: errcheck` only where the error is genuinely unrecoverable (e.g. deferred `Close()`)
- `context.md` is the source of truth for repo structure and data flow — keep it current

---

## File locations (quick reference)

| Concern | Package |
|---|---|
| Domain types | `internal/domain` |
| Config loading + validation | `internal/config` |
| Ingestion | `internal/ingestion` |
| Parsing (CSV / XLSX) | `internal/parsing/{csv,xlsx}` |
| Aggregation + budget groups | `internal/aggregation/household` |
| Report building | `internal/presentation/build.go` |
| Exporters (JSON, Markdown) | `internal/presentation/{json,markdown}` |
| Pipeline orchestration | `internal/pipeline` |
| Manifest | `internal/manifest` |
| Dashboard SPA | `output/index.html` |