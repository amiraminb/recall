# AGENTS: Guidelines for /Users/amir/dev/recall

## Scope
- Applies to the entire repository unless superseded by nested AGENTS.md files.
- No Cursor rules found (`.cursor/rules`, `.cursorrules` absent).
- No Copilot rules found (`.github/copilot-instructions.md` absent).

## Quickstart Summary
- Language: Go (module `github.com/amiraminb/recall`, Go 1.25.3).
- CLI entrypoint: `./cmd/recall` (Cobra-based commands).
- Data persists under `<wiki>/.srs/reviews.json`; config at `~/.config/recall/config.json`.
- Goal: spaced repetition for markdown wikis flagged with `review: true` in YAML frontmatter.

## Build, Run, Lint, Test
- Install: `go install github.com/amiraminb/recall/cmd/recall@latest`.
- Build CLI binary: `go build ./cmd/recall`.
- Build all packages: `go build ./...`.
- Run CLI locally: `go run ./cmd/recall --help`.
- Module tidy (only if deps change): `go mod tidy`.
- Format: `gofmt -w ./` (or `goimports` if already used).
- Lint/basic analysis: `go vet ./...`.
- Test all packages: `go test ./...`.
- Test a single package: `go test ./internal/parser`.
- Test a single test: `go test ./internal/parser -run TestName -count=1`.
- Benchmark (if added): `go test -bench=. ./internal/parser`.
- Prefer focused package-level runs before `./...`.

## Repository Map
- `/cmd/recall`: Cobra commands (`init`, `scan`, `due`, `read`, `review`, `tags`, `history`, `remove`).
- `/internal/config`: config load/save (`~/.config/recall/config.json`).
- `/internal/parser`: markdown frontmatter scanning for `review: true` topics.
- `/internal/storage`: JSON-backed topic/review persistence; FSRS integration.
- `/internal/fsrs`: scheduling logic and rating types (Again/Hard/Good/Easy).
- `README.md`: usage walkthrough and feature descriptions.

## Coding Style: General
- Follow idiomatic Go: short helpers, early returns, clear error paths.
- Keep functions focused; reuse helpers (`getWikiPath`, `getStorage`).
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
- Generate IDs deterministically via `sha256` of file/title (`generateID`).
- FSRS card state managed through `fsrs.Card`; ratings via `fsrs.Rating` enums.
- Maintain backwards compatibility in stored JSON; default missing fields safely.

## Coding Style: File & IO
- Directory permissions `0o755`, file permissions `0o644`.
- Use `os.MkdirAll` before writing config or storage files.
- Use `json.MarshalIndent` with two-space indent for persisted JSON.
- When reading files, call `defer file.Close()` immediately after `os.Open`.
- Skip hidden directories when walking user wikis; only process `.md` files.
- Keep path handling cross-platform via `filepath` utilities.

## Coding Style: Error Handling
- Prefer returning errors from `RunE` and helpers; Cobra prints/propagates.
- Emit user-facing errors to stderr via `fmt.Fprintln(os.Stderr, err)` (see `main.go`).
- Avoid panics; handle expected missing config (`nil, nil` from `config.Load`).
- Provide actionable error messages (e.g., prompt to run `recall init <path>`).
- Wrap context with `fmt.Errorf` when helpful; keep messages concise.
- Error strings should be lowercase with no trailing punctuation.

## Coding Style: CLI UX
- Keep command `Use`, `Short`, `Long` descriptive; include examples in `Long`.
- Register flags in `init()`; prefer simple types (string/bool) for UX.
- Print summaries with `fmt.Printf`; align with existing outputs (`+`, `~`, `?`).
- Avoid noisy logging; favor concise, user-friendly console output.
- Never mutate user wiki files directly (only `.srs` data).

## Coding Style: Collections & Iteration
- Use `slices.Equal` and `slices.Contains` from stdlib for tag comparisons.
- Prefer `range` loops; avoid manual index math unless necessary.
- Build maps with `make(map[string]bool)` when tracking seen items.
- For filtering, allocate new slices rather than modifying in place.

## Coding Style: Time & Scheduling
- Use `time.Now()` captured once per operation.
- Derive `today`/`weekEnd` consistently (`time.Date` truncation to day).
- Express intervals in days via `time.Until(...).Hours() / 24` (see `due`).
- Keep scheduling logic inside FSRS/card helpers when extending behavior.

## Testing Guidance
- New tests should mirror package structure; name files `*_test.go`.
- Use table-driven tests for parser/storage logic.
- For CLI commands, prefer testing underlying helpers over Cobra wiring.
- Avoid global state; create temp dirs for wiki paths/storage files.
- Use `t.Helper()` in shared test helpers.

## Documentation & Comments
- Keep `Long` command descriptions user-focused.
- Avoid inline code comments unless clarifying edge cases.
- Update `README.md` when altering user-facing workflows or storage layout.
- Add doc comments (`// Name ...`) for exported types/functions.

## Dependencies & Versions
- Current deps: Cobra v1.10.2, tablewriter v1.1.2, yaml.v3.
- Avoid adding heavy dependencies; prefer stdlib first.
- If adding a dependency, run `go mod tidy` and justify in PR.

## Storage & Safety
- Never write inside the user wiki except the `.srs` directory.
- Ensure cleanup commands only touch `.srs` data.
- Backwards-compatible migrations: load JSON, add defaults, then save.

## Performance Notes
- Parser uses `bufio.Scanner`; avoid loading entire files into memory.
- Skip hidden directories to keep scans fast; batch writes only if needed.

## Contributing Workflow
- Keep changes minimal and scoped; avoid unrelated refactors.
- Ensure builds/tests pass before requesting review.
- Provide clear commit/PR messages describing behavior changes and reasoning.

## Security & Privacy
- Treat wiki paths as user-sensitive; avoid logging file contents.
- Validate inputs where feasible; avoid executing arbitrary user content.

## Observability
- Prefer deterministic, quiet output; no debug logging by default.
- If adding verbosity, gate behind flags and keep default noise low.

## Extending Commands
- Add new commands as separate files under `cmd/recall`.
- Reuse shared helpers (`getWikiPath`, `getStorage`).
- Keep flag names short and memorable; document in `Long` with examples.

## Frontmatter & Parsing
- `review: true` gates inclusion; respect `id` override, else filename, else heading.
- Preserve tag order when possible; use YAML `tags` array.
- Heading regex: `^(#{1,6})\s+(.+)$`; maintain when extending parsing.

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
