# Devices

This section documents devices that have been tested with sigrok-mcp-server, including connection settings, verified commands, and device-specific notes.

## Tested Devices

| Device | Type | Connection | Notes |
|---|---|---|---|
| [OWON XDM1241](owon-xdm1241.md) | Digital multimeter | USB serial (CH340) | Non-standard SCPI, use `serial_query` |

## Adding device documentation

If you've tested a device with sigrok-mcp-server, contributions are welcome! See [Contributing](../development/contributing.md) for how to submit device documentation.

A device page should include:

- Overview and connection parameters
- Verified commands and expected responses
- Commands that do **not** work (to save others from troubleshooting)
- Usage examples with the appropriate MCP tools
