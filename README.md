# sigrok-mcp-server

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server that wraps [sigrok-cli](https://sigrok.org/wiki/Sigrok-cli), exposing sigrok's signal analysis capabilities to LLMs. It translates MCP tool calls into `sigrok-cli` invocations and returns structured JSON results, enabling LLMs to query logic analyzers, decode protocols, and analyze signals.

## Tools

| Tool | Description |
|---|---|
| `list_supported_hardware` | List all supported hardware drivers |
| `list_supported_decoders` | List all supported protocol decoders |
| `list_input_formats` | List all supported input file formats |
| `list_output_formats` | List all supported output file formats |
| `show_decoder_details` | Show detailed info about a protocol decoder (options, channels, documentation) |
| `show_driver_details` | Show detailed info about a hardware driver (functions, scan options, devices) |
| `show_version` | Show sigrok-cli and library version information |
| `scan_devices` | Scan for connected hardware devices |

## Quickstart

### Docker

```bash
docker build -t sigrok-mcp-server .
docker run -i sigrok-mcp-server
```

### From source

Requires Go 1.25+ and `sigrok-cli` installed on your system.

```bash
go build -o sigrok-mcp-server ./cmd/sigrok-mcp-server
./sigrok-mcp-server
```

The server communicates over stdio (stdin/stdout JSON-RPC).

## Configuration

Configuration is via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SIGROK_CLI_PATH` | `sigrok-cli` | Path to the sigrok-cli binary |
| `SIGROK_TIMEOUT_SECONDS` | `30` | Command execution timeout in seconds |
| `SIGROK_WORKING_DIR` | (empty) | Working directory for sigrok-cli execution |

## MCP Client Configuration

### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "sigrok": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "sigrok-mcp-server"]
    }
  }
}
```

To pass a USB device through to the container:

```json
{
  "mcpServers": {
    "sigrok": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "--privileged", "sigrok-mcp-server"]
    }
  }
}
```

### Claude Code

```bash
claude mcp add sigrok -- docker run -i --rm sigrok-mcp-server
```

## Architecture

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

- **Transport**: stdio (stdin/stdout JSON-RPC)
- **No C bindings**: sigrok-cli is the sole interface to sigrok
- **Read-only**: All tools are read-only queries; no data acquisition or device configuration
- **Structured output**: Raw sigrok-cli text output is parsed into JSON

## Development

```bash
# Build
go build ./...

# Test
go test ./... -race

# Lint
golangci-lint run ./...
```

### Project Structure

```
cmd/sigrok-mcp-server/     Entry point
internal/
  config/                  Environment-based configuration
  sigrok/
    executor.go            sigrok-cli command execution with timeout
    parser.go              Output parsing (list, decoder, driver, version, scan)
    testdata/              Real sigrok-cli output fixtures
  tools/
    tools.go               MCP tool definitions and registration
    handlers.go            Tool handler implementations
```

## License

MIT (Kenos, Inc.)
