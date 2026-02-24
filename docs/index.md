# sigrok-mcp-server

An [MCP (Model Context Protocol)](https://modelcontextprotocol.io/) server that wraps [sigrok-cli](https://sigrok.org/wiki/Sigrok-cli), exposing sigrok's signal analysis capabilities to LLMs.

It translates MCP tool calls into `sigrok-cli` invocations and returns structured JSON results, enabling LLMs to query logic analyzers, decode protocols, and analyze signals.

## What can it do?

- **Discover hardware** — Scan for connected logic analyzers, oscilloscopes, and multimeters
- **Capture data** — Acquire signal data from devices and save to files
- **Decode protocols** — Analyze captured data with 100+ protocol decoders (UART, SPI, I2C, etc.)
- **Query instruments** — Send SCPI commands directly to serial-attached instruments
- **Browse capabilities** — List supported drivers, decoders, and file formats

## Available Tools

| Tool | Description |
|---|---|
| [`list_supported_hardware`](tools/list-supported-hardware.md) | List all supported hardware drivers |
| [`list_supported_decoders`](tools/list-supported-decoders.md) | List all supported protocol decoders |
| [`list_input_formats`](tools/list-input-formats.md) | List all supported input file formats |
| [`list_output_formats`](tools/list-output-formats.md) | List all supported output file formats |
| [`show_decoder_details`](tools/show-decoder-details.md) | Show detailed info about a protocol decoder |
| [`show_driver_details`](tools/show-driver-details.md) | Show detailed info about a hardware driver |
| [`show_version`](tools/show-version.md) | Show sigrok-cli and library version info |
| [`scan_devices`](tools/scan-devices.md) | Scan for connected hardware devices |
| [`capture_data`](tools/capture-data.md) | Capture data from a device and save to file |
| [`decode_protocol`](tools/decode-protocol.md) | Decode protocol data from a captured file |
| [`check_firmware_status`](tools/check-firmware-status.md) | Check firmware file availability |
| [`serial_query`](tools/serial-query.md) | Send commands over a serial port |
| [`get_device_profile`](tools/get-device-profile.md) | Look up device profiles and connection settings |

## Quick Example

A typical signal analysis workflow with an LLM:

1. **Discover** — `scan_devices` to find connected hardware
2. **Inspect** — `show_driver_details` to check driver capabilities
3. **Capture** — `capture_data` to acquire signal data
4. **Decode** — `decode_protocol` to analyze with protocol decoders
5. **Understand** — The LLM interprets decoded output and explains the communication

## Getting Started

See the [Installation](getting-started/installation.md) guide to get up and running.
