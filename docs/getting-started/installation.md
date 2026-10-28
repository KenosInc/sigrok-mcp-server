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

USB passthrough works on **Linux hosts only**. Docker Desktop on macOS and
Windows cannot pass USB devices into containers; use USB/IP or run a Linux
Docker Engine instead.

To access USB-connected devices (logic analyzers, oscilloscopes, etc.) from
inside the container, and to keep captures after the container exits:

```bash
docker run -i --rm --device=/dev/bus/usb -v sigrok-data:/data ghcr.io/kenosinc/sigrok-mcp-server
```

Serial instruments (`serial_query`, `scpi-*` drivers) also need their tty node,
e.g. `--device=/dev/ttyUSB0` or `--device=/dev/ttyACM0`.

!!! note
    `--device=/dev/bus/usb` exposes only the USB bus. With rootful Linux Docker
    the container runs as host root, so no extra udev rule is usually needed;
    with rootless Docker or user namespaces the host device permissions still
    apply and you may need a udev rule granting your user access. Devices are
    bound when the container is created — plug the analyzer in **before**
    starting it. Devices that upload firmware and re-enumerate (e.g. Kingst
    LA2016) need `-v /dev/bus/usb:/dev/bus/usb --device-cgroup-rule='c 189:* rmw'`
    instead of `--device`; see the [firmware guide](firmware.md). Only fall
    back to `--privileged` as a last resort — it grants the container access
    to every host device and all capabilities.
