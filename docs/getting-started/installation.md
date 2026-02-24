# Installation

## Docker (recommended)

Docker is the easiest way to get started — sigrok-cli and all dependencies are bundled in the image.

```bash
docker pull ghcr.io/kenosinc/sigrok-mcp-server
docker run -i --rm ghcr.io/kenosinc/sigrok-mcp-server
```

!!! tip
    You can pin to a specific version tag (e.g. `ghcr.io/kenosinc/sigrok-mcp-server:1.0`) or use `main` for the latest development build.

### Building from Dockerfile

If you prefer to build the image yourself:

```bash
docker build -t sigrok-mcp-server .
docker run -i --rm sigrok-mcp-server
```

## From source

Requires:

- **Go 1.23+**
- **sigrok-cli** installed on your system ([sigrok installation guide](https://sigrok.org/wiki/Downloads))

```bash
go build -o sigrok-mcp-server ./cmd/sigrok-mcp-server
./sigrok-mcp-server
```

## Configuration

Configuration is done via environment variables:

| Variable | Default | Description |
|---|---|---|
| `SIGROK_CLI_PATH` | `sigrok-cli` | Path to the sigrok-cli binary |
| `SIGROK_TIMEOUT_SECONDS` | `30` | Command execution timeout in seconds |
| `SIGROK_WORKING_DIR` | current dir | Working directory for sigrok-cli execution |

For Docker, pass environment variables with `-e`:

```bash
docker run -i --rm -e SIGROK_TIMEOUT_SECONDS=60 ghcr.io/kenosinc/sigrok-mcp-server
```

## USB device access (Docker)

To access USB-connected devices (logic analyzers, oscilloscopes, etc.) from inside the container:

```bash
docker run -i --rm --privileged ghcr.io/kenosinc/sigrok-mcp-server
```

!!! warning
    `--privileged` grants broad access to host devices. For production use, consider using `--device` to pass specific USB devices instead.
