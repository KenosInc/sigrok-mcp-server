# Contributing

Contributions to sigrok-mcp-server are welcome! This guide covers the
development workflow, branching conventions, and code standards.

## Development Setup

### Using the Dev Container (recommended)

The repository includes a dev container configuration. Open the project in
VS Code or any IDE with dev container support, or use:

```bash
devcontainer up
```

### Local Setup

Requirements:

- Go 1.23+
- sigrok-cli (for integration testing)

```bash
# Build
go build ./...

# Test
go test ./... -race

# Lint
go vet ./...
golangci-lint run ./...
```

## Branch Naming

All branches must follow this naming convention:

```
<type>/<short-description>
```

| Prefix | Purpose | Example |
|---|---|---|
| `feat/` | New feature | `feat/serial-port-auto-detect` |
| `fix/` | Bug fix | `fix/timeout-handling` |
| `perf/` | Performance improvement | `perf/parser-allocation` |
| `refactor/` | Code refactoring | `refactor/executor-interface` |
| `docs/` | Documentation only | `docs/troubleshooting-guide` |
| `ci/` | CI/CD changes | `ci/add-lint-workflow` |
| `chore/` | Other maintenance | `chore/update-dependencies` |

Use kebab-case for `<short-description>` (e.g. `feat/serial-port-auto-detect`,
not `feat/serialPortAutoDetect`).

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/) style:

```
<type>: <description>
```

Examples:

```
feat: add serial port auto-detection
fix: handle timeout when device is unresponsive
docs: add OWON XDM1241 device profile guide
ci: add PR labeler workflow
chore: update mcp-go to v0.45.0
perf: reduce allocations in protocol decoder parser
refactor: extract command builder from executor
```

## Pull Requests

1. Create a branch from `main` following the [branch naming](#branch-naming)
   convention
2. Make your changes with tests
3. Ensure `go test ./... -race` passes
4. Ensure `go vet ./...` reports no issues
5. Submit a pull request against `main`

### Automatic Labels

PRs are automatically labeled based on branch name and changed files:

**By branch name:**

| Branch prefix | Label |
|---|---|
| `feat/` | `Type: Feature` |
| `fix/` | `Type: Bug` |
| `perf/` | `Type: Enhancement` |
| `refactor/`, `chore/`, `ci/` | `Type: House Keeping` |
| `docs/` | `Type: Documentation` |

**By changed files:**

| Files | Label |
|---|---|
| `**/*.go`, `go.mod`, `go.sum` | `go` |
| `Dockerfile`, `.devcontainer/**` | `docker` |
| `.github/workflows/**` | `github_actions` |
| `go.mod`, `go.sum` | `dependencies` |

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use table-driven tests with `t.Run()` subtests
- Error handling: return errors, don't panic; wrap with
  `fmt.Errorf("context: %w", err)`
- Naming: exported names in PascalCase, unexported in camelCase
- Keep packages small and focused; avoid circular dependencies

## Areas for Contribution

- **Device documentation:** Test with your hardware and document the results
  in `docs/devices/`
- **Protocol decode examples:** Share real-world decode workflows in
  `docs/guides/`
- **Bug fixes and improvements:** Check
  [open issues](https://github.com/KenosInc/sigrok-mcp-server/issues)
- **Device profiles:** Add JSON profiles for new instruments in
  `internal/devices/`

## Reporting Issues

Please use
[GitHub Issues](https://github.com/KenosInc/sigrok-mcp-server/issues)
to report bugs or request features.
