# Workflow Example

This guide walks through a typical signal analysis session using the MCP tools.

## Scenario

You have a logic analyzer connected via USB and want to capture and decode UART communication.

## Step 1: Scan for devices

First, discover what hardware is connected:

```
Tool: scan_devices
```

The LLM sees the connected logic analyzer and its driver ID.

## Step 2: Check driver capabilities

Inspect the driver to understand configuration options:

```
Tool: show_driver_details
Parameters:
  driver: fx2lafw
```

This reveals supported sample rates, channel counts, and other capabilities.

## Step 3: Capture data

Capture signal data from the device:

```
Tool: capture_data
Parameters:
  driver: fx2lafw
  channels: D0
  config: samplerate=1M
  samples: 1000000
```

The captured data is saved to a file.

## Step 4: Decode the protocol

Apply a UART decoder to the captured data:

```
Tool: decode_protocol
Parameters:
  input_file: capture_20260224_120000.sr
  protocol_decoders: uart:baudrate=9600:rx=D0
```

## Step 5: LLM analysis

The LLM receives the decoded output — individual bytes, frames, and timing information — and can:

- Explain what the communication means
- Identify protocol errors or anomalies
- Suggest next steps (different baud rate, additional channels, stacked decoders)

## Working with serial instruments

For SCPI-capable instruments (multimeters, power supplies, etc.), you can communicate directly:

```
Tool: get_device_profile
Parameters:
  query: XDM1241
```

Then query the instrument using the profile's settings:

```
Tool: serial_query
Parameters:
  port: /dev/ttyUSB0
  command: "MEAS?"
  baudrate: 115200
  timeout_ms: 3000
```
