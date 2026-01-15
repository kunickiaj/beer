# AGENTS

## Scope
This file applies to the entire repository.

## Project Overview
- Go CLI for managing JIRA/Gerrit/git workflow.
- Configuration lives at `~/.beer.yaml` by default (override with `--config`).
- Key commands: `beer brew`, `beer taste` (see `README.md`).
- CI runs Go `^1.25` and `golangci-lint`.

## Build, Lint, Test
Run these from repo root unless noted.

### Dependencies
- `go mod download`

### Build
- `go build -v ./...`

### Lint
- `golangci-lint run ./...`

### Test
- `go test -v ./...`

### Single Test (fast feedback)
- Run a specific test by name: `go test ./... -run TestName -count=1`
- Run a specific package: `go test ./cmd -run TestName -count=1`
- Short mode (if tests honor it): `go test ./... -short`

### Formatting
- `gofmt -w <files>` (always run on modified Go files)

### Release (optional)
- Dry run: `make release-dry-run`
- Publish: `make release` (requires `.release-env`)

## Code Style Guidelines
These mirror the project standards and existing code.

### General Principles
- Keep changes small and focused.
- Favor pure, testable functions; refactor if tests are hard.
- Use explicit dependencies; avoid hidden globals.
- Prefer declarative logic over nested imperative loops.

### Go Conventions
- Follow standard Go idioms and `gofmt` formatting.
- Use `goimports` if available to keep imports tidy.
- Keep functions small (ideally < 50 lines).
- Use early returns to avoid deep nesting.
- Avoid unnecessary interfaces; accept concrete types unless needed.

### Imports
- Group standard library, third-party, and local imports.
- Keep imports alphabetical within groups.
- Remove unused imports before committing.

### Naming
- Use Go naming conventions: `camelCase` for locals, `PascalCase` for exported.
- Prefer descriptive names: `configPath` over `cp`.
- Predicates should read like booleans: `isDraft`, `hasConfig`.
- Constants use `lowerCamel` unless exported.

### Types and Data
- Favor explicit types in public APIs and interfaces.
- Keep structs focused; avoid god-structs.
- Prefer value types unless mutation is required.

### Error Handling
- Handle errors explicitly; do not ignore return values.
- Add context with `fmt.Errorf("...: %w", err)` when returning.
- Avoid `panic` in CLI flow; return errors or `log.Fatal` with context.
- Don’t leak secrets in logs or error messages.

### Logging
- Use `logrus` as seen in existing code.
- Include context fields when helpful (e.g., IDs, config keys).
- Debug logs are fine; avoid noisy info logs by default.

### Config & Security
- Keep external integrations optional/configurable.
- Never hardcode secrets or tokens.
- Validate inputs at boundaries; sanitize user-provided strings.
- Be cautious when reading config defaults.

### CLI UX
- Keep command descriptions short and actionable.
- Errors should be clear and actionable.
- Respect `--dry-run` semantics where present.

### Testing Guidance
- Use AAA pattern (Arrange → Act → Assert).
- Test behavior, not implementation.
- Include happy path, edge cases, and error cases.
- Prefer deterministic tests with minimal setup.

### Docs
- Update `README.md` for user-facing changes.
- Document the “why” for non-obvious behavior.

## Repo-Specific Notes
- No Cursor or Copilot rules found in `.cursor/rules/`, `.cursorrules`, or `.github/copilot-instructions.md`.
- Config file defaults are user-facing; avoid changing them unless required.

## Handy References
- `README.md` for configuration and common workflow.
- `.golangci.yml` for lint defaults.
- `.github/workflows/pr.yaml` for CI build/test commands.

_(If in doubt, gofmt. It’s the tabular path to enlightenment.)_
