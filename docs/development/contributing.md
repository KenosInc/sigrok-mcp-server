# Contributing

Contributions to sigrok-mcp-server are welcome!

## Development setup

### Using the dev container (recommended)

The repository includes a dev container configuration. Open the project in VS Code or any IDE with dev container support, or use:

```bash
devcontainer up
```

### Local setup

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

## Code style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use table-driven tests with `t.Run()` subtests
- Error handling: return errors, don't panic; wrap with `fmt.Errorf("context: %w", err)`
- Naming: exported names in PascalCase, unexported in camelCase

## Pull requests

1. Fork the repository and create a feature branch
2. Make your changes with tests
3. Ensure `go test ./... -race` passes
4. Ensure `go vet ./...` reports no issues
5. Submit a pull request against `main`

## Areas for contribution

- **Device documentation:** Test with your hardware and document the results in `docs/devices/`
- **Protocol decode examples:** Share real-world decode workflows in `docs/guides/`
- **Bug fixes and improvements:** Check [open issues](https://github.com/KenosInc/sigrok-mcp-server/issues)
- **Device profiles:** Add JSON profiles for new instruments in `internal/devices/`

## Reporting issues

Please use [GitHub Issues](https://github.com/KenosInc/sigrok-mcp-server/issues) to report bugs or request features.
