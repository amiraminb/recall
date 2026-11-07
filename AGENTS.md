# AGENTS: Guidelines for /Users/amir/dev/recall

## Scope
- Applies to the entire repository unless superseded by nested AGENTS.md files.
- No Cursor rules found (`.cursor/rules`, `.cursorrules` absent).
- No Copilot rules found (`.github/copilot-instructions.md` absent).

## Quickstart Summary
- Language: Go (module `github.com/amiraminb/recall`, Go 1.25.6).
- CLI entrypoint: `./cmd/recall` (Cobra-based commands).
- Data persists under `$XDG_DATA_HOME/recall/reviews.json` or `~/.local/share/recall/reviews.json`.
- Goal: spaced repetition for topics users explicitly add after studying them.

## Build, Run, Lint, Test
- Install: `go install github.com/amiraminb/recall/cmd/recall@latest`.
- Build CLI binary: `go build ./cmd/recall`.
- Build all packages: `go build ./...`.
- Run CLI locally: `go run ./cmd/recall --help`.
- Module tidy (only if deps change): `go mod tidy`.
- Format: `gofmt -w ./` (or `goimports` if already used).
- Lint/basic analysis: `go vet ./...`.
- Test all packages: `go test ./...`.
- Test a single package: `go test ./internal/storage`.
- Test a single test: `go test ./internal/storage -run TestName -count=1`.
- Benchmark (if added): `go test -bench=. ./internal/storage`.
- Prefer focused package-level runs before `./...`.

## Repository Map
- `/cmd/recall`: Cobra commands (`add`, `due`, `review`, `list`, `edit`, `open`, `tags`, `history`, `remove`).
- `/internal/config`: standalone data directory and legacy config loading for migration.
- `/internal/storage`: JSON-backed topic/review persistence; FSRS integration.
- `/internal/fsrs`: scheduling logic and rating types (Again/Hard/Good/Easy).
- `README.md`: usage walkthrough and feature descriptions.

## Coding Style: General
- Follow idiomatic Go: short helpers, early returns, clear error paths.
- Keep functions focused; reuse helpers (`getStorage`, `resolveTopic`).
- Prefer pure functions; inject dependencies instead of globals.
- Avoid one-letter names except conventional loops (`i`, `j`).
- Keep boolean flags descriptive (`week`, `tag`, `review`).

## Coding Style: Naming
- Exported identifiers in TitleCase; unexported in camelCase.
- Acronyms follow Go style (`ID`, `URL`, `FSRS`).
- Avoid stutter in type names (`storage.Store`, not `storage.StorageStore`).

## Coding Style: Imports & Modules
- Standard library imports first; third-party imports after a blank line.
- Keep command files focused per verb in `cmd/recall`.
- Register Cobra commands in `init()`; use `rootCmd.AddCommand`.
- Avoid cyclic dependencies; pass data instead of shared globals.

## Coding Style: Types & Data
- Storage models live in `internal/storage/models.go`.
- Keep JSON tags in snake_case.
- Use `time.Time` for timestamps; store with `ReviewedAt` and `Created` fields.
- Generate IDs via `generateID`; new topics use creation time/title, while migrated IDs remain stable.
- FSRS card state managed through `fsrs.Card`; ratings via `fsrs.Rating` enums.
- Maintain backwards compatibility in stored JSON; default missing fields safely.

## Coding Style: File & IO
- Directory permissions `0o755`, file permissions `0o644`.
- Use `os.MkdirAll` before writing config or storage files.
- Use `json.MarshalIndent` with two-space indent for persisted JSON.
- When reading files, call `defer file.Close()` immediately after `os.Open`.
- Source references are metadata; do not scan directories or parse user notes.
- Keep path handling cross-platform via `filepath` utilities.

## Coding Style: Error Handling
- Prefer returning errors from `RunE` and helpers; Cobra prints/propagates.
- Emit user-facing errors to stderr via `fmt.Fprintln(os.Stderr, err)` (see `main.go`).
- Avoid panics; handle expected missing config (`nil, nil` from `config.Load`).
- Provide actionable error messages (e.g., prompt to run `recall add "Topic"`).
- Wrap context with `fmt.Errorf` when helpful; keep messages concise.
- Error strings should be lowercase with no trailing punctuation.

## Coding Style: CLI UX
- Keep command `Use`, `Short`, `Long` descriptive; include examples in `Long`.
- Register flags in `init()`; prefer simple types (string/bool) for UX.
- Print summaries through `cmd.OutOrStdout()` so output can be redirected and tested.
- Avoid noisy logging; favor concise, user-friendly console output.
- Never mutate user source materials or legacy wiki data.

## Coding Style: Collections & Iteration
- Use `slices.Equal` and `slices.Contains` from stdlib for tag comparisons.
- Prefer `range` loops; avoid manual index math unless necessary.
- Build maps with `make(map[string]bool)` when tracking seen items.
- For filtering, allocate new slices rather than modifying in place.

## Coding Style: Time & Scheduling
- Use `time.Now()` captured once per operation.
- Derive `today`/`weekEnd` consistently (`time.Date` truncation to day).
- Use local calendar days for the daily queue and due-date labels.
- Keep scheduling logic inside FSRS/card helpers when extending behavior.

## Testing Guidance
- New tests should mirror package structure; name files `*_test.go`.
- Use table-driven tests for storage and review-session logic.
- For CLI commands, prefer testing underlying helpers over Cobra wiring.
- Avoid global state; create temp dirs for storage files and legacy migration fixtures.
- Use `t.Helper()` in shared test helpers.

## Documentation & Comments
- Keep `Long` command descriptions user-focused.
- Avoid inline code comments unless clarifying edge cases.
- Update `README.md` when altering user-facing workflows or storage layout.
- Add doc comments (`// Name ...`) for exported types/functions.

## Dependencies & Versions
- Current deps: Cobra v1.10.2, tablewriter v1.1.2, fatih/color v1.15.0.
- Avoid adding heavy dependencies; prefer stdlib first.
- If adding a dependency, run `go mod tidy` and justify in PR.

## Storage & Safety
- Write only to Recall's standalone data directory.
- Legacy migration reads `~/.config/recall/config.json` and copies `.srs/reviews.json` only when standalone data is absent.
- Preserve legacy IDs, schedules, and history; never overwrite existing standalone data during migration.

## Performance Notes
- Use one buffered reader per interactive session so piped input is not lost.
- Persist each review's schedule and log together before advancing the queue.

## Contributing Workflow
- Keep changes minimal and scoped; avoid unrelated refactors.
- Ensure builds/tests pass before requesting review.
- Provide clear commit/PR messages describing behavior changes and reasoning.

## Security & Privacy
- Treat topics and source references as user-sensitive; avoid logging data by default.
- Validate inputs where feasible; avoid executing arbitrary user content.

## Observability
- Prefer deterministic, quiet output; no debug logging by default.
- If adding verbosity, gate behind flags and keep default noise low.

## Extending Commands
- Add new commands as separate files under `cmd/recall`.
- Reuse shared helpers (`getStorage`, `resolveTopic`).
- Keep flag names short and memorable; document in `Long` with examples.

## Topics & Review Flow
- A title is required; source and tags are optional.
- Adding means the topic has been studied; schedule tomorrow without inventing a recall rating.
- Encourage recall before checking the source; use FSRS for actual review ratings.
- Match titles case-insensitively and preserve tag order when possible.

## Release & Distribution
- Preferred distribution via `go install ...@latest`.
- Keep `go.mod` tidy for module consumers.
- Maintain backward compatibility for CLI flags where possible.

## Miscellaneous
- No autogenerated code present; keep files small and cohesive.
- Avoid vendoring dependencies unless absolutely required.
- Keep line lengths reasonable; gofmt handles wrapping.

## Contact
- For missing guidance, follow idiomatic Go and mirror nearby patterns.
