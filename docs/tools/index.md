# Tools Overview

sigrok-mcp-server exposes the following MCP tools. Each tool translates into one or more `sigrok-cli` invocations and returns structured JSON results.

## Discovery tools

These tools query sigrok's capabilities and require no connected hardware.

| Tool | Description |
|---|---|
| [`list_supported_hardware`](list-supported-hardware.md) | List all supported hardware drivers |
| [`list_supported_decoders`](list-supported-decoders.md) | List all supported protocol decoders |
| [`list_input_formats`](list-input-formats.md) | List all supported input file formats |
| [`list_output_formats`](list-output-formats.md) | List all supported output file formats |
| [`show_decoder_details`](show-decoder-details.md) | Show detailed info about a protocol decoder |
| [`show_driver_details`](show-driver-details.md) | Show detailed info about a hardware driver |
| [`show_version`](show-version.md) | Show sigrok-cli and library version info |

## Device tools

These tools interact with connected hardware.

| Tool | Description |
|---|---|
| [`scan_devices`](scan-devices.md) | Scan for connected hardware devices |
| [`capture_data`](capture-data.md) | Capture data from a device and save to file |
| [`check_firmware_status`](check-firmware-status.md) | Check firmware file availability |

## Analysis tools

These tools process captured data.

| Tool | Description |
|---|---|
| [`decode_protocol`](decode-protocol.md) | Decode protocol data from a captured file |

## Instrument tools

These tools communicate directly with serial-attached instruments, independent of sigrok-cli.

| Tool | Description |
|---|---|
| [`serial_query`](serial-query.md) | Send commands over a serial port |
| [`get_device_profile`](get-device-profile.md) | Look up device profiles and connection settings |
