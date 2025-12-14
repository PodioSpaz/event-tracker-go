# Repository Guidelines

## Project Structure & Module Organization
- Single Go module targeting Go 1.23 via `go.mod`.
- `cmd/app` builds the Fyne desktop binary; `cmd/migrate` powers the TinyDB importer.
- `internal/` holds domain, repository, service, controller, UI, config, and util packages; treat them as module-private code.
- `pkg/migration` exposes reusable import pieces, SQL scripts live in `migrations/`, and runtime artifacts sit in `data/` (SQLite) plus `logs/`.
- Tests mirror the package layout with fixtures in `testdata/`.

## Build, Test, and Development Commands
- `task dev-setup`: download modules and create writable `data/` and `logs/`.
- `task build`: compile `bin/event-tracker` and `bin/migrate`.
- `task run`: start the GUI using `go run ./cmd/app`.
- `task migrate`: ensure the migrate CLI exists and run the latest SQL migrations into `data/events.db`.
- `task test` (race + coverage) or `task test-short`: run tests and refresh `coverage.out`.
- `task check`: run `fmt`, `vet`, and the test suite before opening a pull request.

## Coding Style & Naming Conventions
- Follow gofmt defaults: tabs, `camelCase` locals, exported `PascalCase` APIs, and `_test.go` suffixes.
- Run `task fmt`, `task vet`, or `task lint` before commits to catch formatting and vet issues.
- UI structs under `internal/ui` typically use a `FooView` suffix, and SQLite-specific files can adopt a `sqlite_*.go` prefix to signal backend coupling.

## Testing Guidelines
- Use the standard `testing` package with table-driven cases.
- Keep fast unit tests near their packages; save heavier SQLite integration checks for repository packages and seed them with `testdata/`.
- Always run `task test` so race detection and coverage stay current; note big coverage swings when touching services or controllers.

## Commit & Pull Request Guidelines
- Commits follow Conventional Commits (`feat:`, `fix:`, `chore:`) as shown in `feat: initial Event Tracker Go implementation`; keep subjects under ~70 characters.
- PR descriptions should summarize scope, note schema or migration impacts, add manual test notes, and link roadmap items.
- Attach screenshots or GIFs for Fyne changes and confirm `task check` before requesting review; document any failing step you could not reproduce.

## Security & Configuration Tips
- SQLite databases in `data/` often contain PII; clean them with `task clean` or `task db-reset` before sharing builds.
- Viper reads env vars and optional config files—store secrets outside the repo.
- Logs can leak attendee emails, so scrub `logs/` before uploading diagnostics, and verify importer samples with `bin/migrate -cmd verify -source <file>`.
