# AGENTS: Guidelines for /Users/amir/dev/recall

## Scope
- Applies to the entire repository unless superseded by future nested AGENTS.md files.
- No Cursor rules found (.cursor/rules, .cursorrules absent).
- No Copilot rules found (.github/copilot-instructions.md absent).

## Quickstart Summary
- Language: Go (module github.com/amiraminb/recall, go 1.25.3).
- CLI entrypoint: `./cmd/recall` (Cobra-based commands).
- Data persists under `<wiki>/.srs/reviews.json`; config at `~/.config/recall/config.json`.
- Goal: spaced repetition for markdown wikis flagged with `review: true` in YAML frontmatter.

## Build & Run
- Install: `go install github.com/amiraminb/recall/cmd/recall@latest`.
- Build binary: `go build ./cmd/recall` (produces `recall`).
- Build all packages: `go build ./...`.
- Run CLI locally: `go run ./cmd/recall --help`.
- Example workflow: `recall init <path>` → `recall scan` → `recall due`/`recall review`.
- Module tidy (if deps change): `go mod tidy`.

## Testing
- Current repository has no *_test.go files; add tests nearby when creating new features.
- Run full suite: `go test ./...`.
- Run a single test: `go test ./path/to/pkg -run TestName -count=1`.
- Benchmark (if added): `go test -bench=. ./path/to/pkg`.
- Prefer focused package-level runs before `./...`.

## Lint & Formatting
- Use `gofmt -w ./` or goimports before committing changes.
- Basic static analysis: `go vet ./...`.
- Keep imports grouped: standard library first, blank line, third-party modules.
- Avoid introducing new lint tools unless project adopts them.
- Keep error strings lowercase, no trailing punctuation.

## Repository Map
- `/cmd/recall`: Cobra commands (`init`, `scan`, `due`, `read`, `review`, `tags`, `history`, `remove`).
- `/internal/config`: config load/save (`~/.config/recall/config.json`).
- `/internal/parser`: markdown frontmatter scanning for `review: true` topics.
- `/internal/storage`: JSON-backed topic/review persistence; FSRS integration.
- `/internal/fsrs`: scheduling logic and rating types (Again/Hard/Good/Easy).
- Root `README.md`: usage walkthrough and feature descriptions.

## Coding Style: General
- Follow idiomatic Go: short helper functions, early returns on errors.
- Use meaningful names; exported identifiers are TitleCase; unexported are camelCase.
- Avoid one-letter names except short loops (`i`, `t`, `r`) where conventional.
- Keep functions focused; reuse helper utilities (e.g., `getWikiPath`, `getStorage`).
- Prefer pure functions where possible; inject dependencies via parameters rather than globals.
- Keep boolean flags descriptive (`week`, `tag`, `review`).

## Coding Style: Imports & Modules
- Standard library imports first; third-party imports after a blank line.
- Cobra commands live in `cmd/recall` and register in `init()`; keep command files focused per verb.
- Avoid cyclic dependencies; prefer passing data instead of shared globals.

## Coding Style: Types & Data
- Storage models in `internal/storage/models.go`; keep JSON tags in snake_case.
- Use `time.Time` for timestamps; store with `ReviewedAt` and `Created` fields.
- Generate IDs deterministically via `sha256` of file/title (`generateID`).
- FSRS card state managed through `fsrs.Card`; ratings via `fsrs.Rating` enums.
- When extending data, maintain backward compatibility of stored JSON; default missing fields safely.

## Coding Style: File & IO
- File permissions: directories `0o755`, files `0o644` per existing code.
- Use `os.MkdirAll` before writing config or storage files.
- Use `json.MarshalIndent` with two-space indent for persisted JSON.
- When reading files, close handles with `defer file.Close()` immediately after `os.Open`.
- Skip hidden directories when walking user wikis; only process `.md` files.
- Keep path handling cross-platform via `filepath` utilities.

## Coding Style: Error Handling
- Prefer returning errors from `RunE` and helpers; Cobra prints/propagates.
- Emit user-facing errors to stderr via `fmt.Fprintln(os.Stderr, err)` (see `main.go`).
- Avoid panics; handle expected missing config (`nil, nil` from `config.Load`).
- Provide actionable error messages (e.g., prompt to run `recall init <path>` when wiki missing).
- Wrap contextual errors with `fmt.Errorf` when helpful; keep messages concise.

## Coding Style: CLI UX
- Keep command `Use`, `Short`, `Long` descriptive; include examples in `Long` where helpful.
- Register flags in `init()`; prefer simple types (string/bool) for UX.
- Print summaries with `fmt.Printf`, align with existing outputs (`+`, `~`, `?` markers in scan).
- Avoid noisy logging; favor concise, user-friendly console output.
- Respect dry runs if added; never mutate user wiki files directly (only `.srs` data).

## Coding Style: Collections & Iteration
- Use `slices.Equal` and `slices.Contains` from stdlib for tag comparisons.
- Prefer `range` loops; avoid manual index math unless necessary.
- Build maps with `make(map[string]bool)` when tracking seen items.
- For filtering, allocate new slices rather than modifying in place when clarity wins.

## Coding Style: Time & Scheduling
- Use `time.Now()` captured once per operation; derive `today`/`weekEnd` consistently (`time.Date` truncation to day).
- Express intervals in days via `time.Until(...).Hours() / 24` as shown in `due` command.
- Keep all scheduling logic inside FSRS/card helpers when extending algorithm behavior.

## Testing Guidance
- New tests should mirror package structure; name files `*_test.go` adjacent to code.
- Use table-driven tests for parser/storage logic; cover YAML parsing, orphan detection, tag updates.
- For CLI commands, prefer testing underlying helpers over Cobra wiring unless necessary.
- Avoid global state in tests; create temp dirs for wiki paths and storage files.
- Use `t.Helper()` in shared test helpers.

## Documentation & Comments
- Keep `Long` command descriptions user-focused; avoid inline code comments unless clarifying edge cases.
- Update `README.md` when altering user-facing workflows or storage layout.
- Add doc comments (`// Name ...`) for exported types/functions.

## Dependencies & Versions
- Current deps: Cobra v1.10.2, tablewriter v1.1.2, yaml.v3.
- Avoid adding heavy dependencies; prefer stdlib first.
- If adding a dependency, run `go mod tidy` and justify in PR.

## Storage & Safety
- Never write inside the user wiki except `.srs` directory; avoid mutating note contents.
- Ensure `recall remove` or similar cleanup functions only touch `.srs` data.
- Backwards-compatible migrations: load existing JSON, add defaults, then save.

## Performance Notes
- Parser uses `bufio.Scanner`; avoid loading entire files into memory unnecessarily.
- Skip hidden directories to keep scans fast; consider batching writes only if needed.

## Contributing Workflow
- Keep changes minimal and scoped; avoid unrelated refactors.
- Ensure builds/tests pass before requesting review.
- Provide clear commit/PR messages describing behavior changes and reasoning.

## Security & Privacy
- Treat wiki paths as user-sensitive; do not log contents or file paths beyond necessary outputs.
- Validate inputs where feasible; avoid executing arbitrary user content.

## Observability
- Prefer deterministic, quiet output; no debug logging by default.
- If adding verbosity, gate behind flags and keep default noise low.

## Extending Commands
- Add new commands as separate files under `cmd/recall`; register with `rootCmd.AddCommand` in `init()`.
- Reuse shared helpers (`getWikiPath`, `getStorage`) to reduce duplication.
- Keep flag names short and memorable; document in `Long` text with examples.

## Frontmatter & Parsing
- `review: true` gates inclusion; respect `id` override, else filename, else first heading.
- Preserve tag order when possible; use YAML `tags` array.
- Regex for headings: `^(#{1,6})\s+(.+)$`; maintain when extending heading parsing.

## Release & Distribution
- Preferred distribution via `go install ...@latest`; ensure `go.mod` stays tidy for module consumers.
- Maintain backward compatibility for CLI flags where possible.

## Miscellaneous
- No autogenerated code present; keep files small and cohesive.
- Avoid vendoring dependencies unless absolutely required.
- Keep line lengths reasonable; gofmt handles wrapping.

## Contact
- For missing guidance, follow idiomatic Go conventions and mirror existing patterns in neighboring files.
