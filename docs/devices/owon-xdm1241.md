# OWON XDM1241 SCPI Reference

## Overview

The OWON XDM1241 is a 4 1/2 digit bench-type digital multimeter with USB serial connectivity. It is a variant of the XDM1041 (battery-powered, USB-C powered instead of AC mains). Internally it uses a CH340 USB-to-serial chip.

sigrok's `scpi-dmm` driver does not recognize this device. Use the `serial_query` MCP tool for direct SCPI communication.

## Connection Parameters

| Parameter | Value |
|---|---|
| Port | `/dev/ttyUSB0` (typical) |
| Baud rate | 115200 |
| Data bits | 8 |
| Parity | None |
| Stop bits | 1 |
| Timeout | 3000 ms (recommended) |

## Command Conventions

The XDM1241 uses a **non-standard, flat SCPI command set**. Key differences from standard SCPI:

- **No subsystem hierarchy**: Commands like `MEAS:VOLT:DC?`, `VOLT:RANG?`, or `SYST:ERR?` are not supported and will time out with no response.
- **Display suffix**: Many commands use a numeric suffix (`1` = main display, `2` = sub display). Commands without a suffix default to the main display.
- **Write responses**: Non-query commands (those not ending in `?`) return `OK\n` or `nOK\n` (firmware v25052021+).
- **Line terminator**: Responses are terminated with `\n`. Commands should be sent with `\n`.

## Verified Commands

Tested on firmware V4.3.0 via `serial_query` MCP tool.

### Device Identification

| Command | Example Response | Description |
|---|---|---|
| `*IDN?` | `OWON,XDM1241,24412417,V4.3.0,3` | Device identification (manufacturer, model, serial, firmware, unknown) |

### Measurement

| Command | Example Response | Description |
|---|---|---|
| `MEAS?` | `-2.758856E-05` | Read main display value (shorthand for `MEAS1?`) |
| `MEAS1?` | `-2.758856E-05` | Read main display value |
| `MEAS2?` | `NONe` | Read sub display value (`NONe` if sub display is off) |

### Function / Mode

| Command | Example Response | Description |
|---|---|---|
| `FUNC?` | `"VOLT"` | Current function on main display |
| `FUNC1?` | `"VOLT"` | Current function on main display (explicit) |
| `FUNC2?` | `NONe` | Current function on sub display |

### Range

| Command | Example Response | Description |
|---|---|---|
| `RANGE1?` | `50 mV` | Current range on main display |
| `AUTO1?` | `1` | Auto-range status on main display (1=on, 0=off) |

### Sampling Rate

| Command | Example Response | Description |
|---|---|---|
| `RATE?` | `S` | Sampling rate (`S`=Slow, `M`=Medium, `F`=Fast) |

## Commands That Do NOT Work

These standard SCPI commands time out with no response:

- `MEAS:VOLT:DC?` (standard SCPI measurement)
- `VOLT:RANG?`, `VOLT:RANG:AUTO?`
- `SYST:ERR?`
- `CONF?`
- `RANG?`
- `*RST` (no response; may or may not execute)
- `VAL1?`, `VAL2?`
- `RANGE2?` (when sub display is off)

## Usage Examples

### Read voltage with serial_query

```
Tool: serial_query
Parameters:
  port: /dev/ttyUSB0
  command: MEAS1?
  baudrate: 115200
  timeout_ms: 3000
```

Response: `{"response": "-2.758856E-05"}`

### Identify the device

```
Tool: serial_query
Parameters:
  port: /dev/ttyUSB0
  command: *IDN?
  baudrate: 115200
  timeout_ms: 2000
```

Response: `{"response": "OWON,XDM1241,24412417,V4.3.0,3"}`

## References

- [TheHWcave/OWON-XDM1041](https://github.com/TheHWcave/OWON-XDM1041) - Community SCPI command list and firmware info
- [Elektroarzt/owon-xdm-remote](https://github.com/Elektroarzt/owon-xdm-remote) - ESP32 remote control using SCPI over WiFi/MQTT
- [OWON XDM1000 Programming Manual](https://files.owon.com.cn/software/Application/XDM1000_Digital_Multimeter_Programming_Manual.pdf) - Official (incomplete) documentation
- [Owon XDM2041 - sigrok wiki](https://sigrok.org/wiki/Owon_XDM2041) - Related model in sigrok
