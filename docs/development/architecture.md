# Architecture

## Overview

```
MCP Client (LLM)
    |  stdio (JSON-RPC)
    v
sigrok-mcp-server (Go)
    |  exec.Command
    v
sigrok-cli
    |
    v
libsigrok / libsigrokdecode
```

- **Transport:** stdio (stdin/stdout JSON-RPC)
- **No C bindings:** sigrok-cli is the sole interface to sigrok
- **Structured output:** Raw sigrok-cli text output is parsed into JSON

## Package structure

```
cmd/sigrok-mcp-server/     Entry point
internal/
  config/                  Environment-based configuration
  sigrok/
    executor.go            sigrok-cli command execution with timeout
    parser.go              Output parsing (list, decoder, driver, version, scan)
    testdata/              Real sigrok-cli output fixtures
  serial/                  Serial port querier (independent of sigrok-cli)
  devices/                 Device profile registry with embedded JSON profiles
  tools/
    tools.go               MCP tool definitions and registration
    handlers.go            Tool handler implementations
```

## Key design decisions

### sigrok-cli as the sole interface

The server does not use C bindings or link against libsigrok directly. Instead, it shells out to `sigrok-cli` for all sigrok operations. This simplifies the build (pure Go binary), avoids CGO complexity, and leverages the same CLI tool that sigrok users already know.

### Command injection prevention

All user-provided inputs are validated with regexes before being passed to `sigrok-cli`. Filenames are restricted to flat names (no path separators) to prevent path traversal.

### Testability

- `Runner` interface in `internal/tools/` enables mock-based handler tests
- `CommandFactory` in `internal/sigrok/executor.go` is a test seam for injecting fake commands
- Golden output files in `internal/sigrok/testdata/` enable parser tests without sigrok-cli

### serial_query independence

The `serial_query` and `get_device_profile` tools operate independently of sigrok-cli, using a pure Go serial library. This allows instrument communication even when sigrok's SCPI driver doesn't support the device.
