package tools

// Instructions is sent to LLM clients as part of the MCP initialize response.
// It provides guidance on how to use this server's tools effectively.
const Instructions = `# sigrok MCP Server

## Recommended Workflow

1. **Identify the device**: Use ` + "`scan_devices`" + ` to discover connected hardware.
   If a device is not found, use ` + "`check_firmware_status`" + ` to diagnose missing firmware.
2. **Get device details**: Use ` + "`show_driver_details`" + ` to see driver capabilities,
   or ` + "`get_device_profile`" + ` + ` + "`serial_query`" + ` for serial instruments.
3. **Capture data**: Use ` + "`capture_data`" + ` to acquire signals (see Capture Timing Guide below).
4. **Decode protocols**: Use ` + "`decode_protocol`" + ` on the captured file to extract protocol-level data.
   Use ` + "`show_decoder_details`" + ` to check available options and channel mappings before decoding.

## Serial / SCPI Instruments

This server supports direct serial communication with SCPI and SCPI-like instruments
via ` + "`serial_query`" + `. Many bench instruments (multimeters, power supplies, oscilloscopes)
accept text commands over serial ports.

### Device profiles as resources

Device profiles are available as MCP resources (` + "`device://{profile_id}`" + `).
Each profile contains connection settings (baudrate, parity, etc.), supported commands
with example responses, and device-specific notes.

**Always check the device profile before sending commands** — some instruments
(e.g. OWON XDM1241) use non-standard SCPI dialects where standard commands
like ` + "`MEAS:VOLT:DC?`" + ` do not work. The profile documents the correct command set.

### Typical workflow for serial instruments

1. Use ` + "`get_device_profile`" + ` with the device name or ` + "`*IDN?`" + ` response to look up the profile
2. Use ` + "`serial_query`" + ` with ` + "`port`" + `, ` + "`command`" + `, and the connection settings from the profile
3. If the device is unknown, try ` + "`serial_query`" + ` with ` + "`*IDN?`" + ` first, then search for a matching profile

### SCPI vs non-standard instruments

- **Standard SCPI**: Uses hierarchical commands (e.g. ` + "`MEAS:VOLT:DC?`" + `, ` + "`SYST:ERR?`" + `)
- **Non-standard (SCPI-like)**: Some instruments implement a flat or proprietary command set
  that superficially resembles SCPI but is incompatible. Device profiles document these differences.

## Capture Timing Guide

### Synchronous vs Asynchronous Capture

` + "`capture_data`" + ` blocks until the capture window completes. Choose between synchronous
and asynchronous approaches depending on whether you control when the target signal occurs.

#### Synchronous Capture (signal timing is controllable)

Use when the client can control when the signal occurs
(e.g. calling another MCP tool, executing an external command).

**Pattern A: Subagent approach**
- Subagent (background): run ` + "`capture_data`" + ` with a long window (e.g. 15s)
- Main agent: call the target tool repeatedly while capture is running
- Subagent: return decoded results after capture completes

**Pattern B: User-coordinated approach**
Use when the user manually triggers the signal:
1. Inform the user that capture is ready (do not start yet)
2. User gives the go-ahead
3. Start capture
4. User triggers the signal and reports completion
5. Decode and analyze the capture result

#### Asynchronous Capture (signal timing is not controllable)

Use when the signal timing is unpredictable (e.g. responses from external devices,
interrupt signals).

- Set a sufficiently long ` + "`time`" + ` parameter to ensure the signal falls within the window
- Use ` + "`triggers`" + ` with ` + "`wait_trigger: true`" + ` to capture on an edge when possible
- Be aware of timeouts when using triggers (exceeding the deadline causes an error)

### Important notes

- Calling ` + "`capture_data`" + ` in parallel with other MCP tools does not guarantee execution
  order. If the capture completes before the signal occurs, the result will be empty
- MCP servers cannot spawn subagents internally.
  All orchestration must be done on the client side`
