package tools

// Instructions is sent to LLM clients as part of the MCP initialize response.
// It provides guidance on how to use this server's tools effectively.
const Instructions = `# sigrok MCP Server — Usage Guide

## Device Discovery

1. Call scan_devices to find connected hardware. Never guess driver names —
   they are often non-obvious (e.g. Kingst LA2016 uses "kingst-la2016", not "fx2lafw").
2. If scan_devices returns empty, call check_firmware_status — USB analyzers
   fail silently without firmware files.
3. Call show_driver_details to check supported config options before capturing.

## Capturing Data

capture_data blocks the entire MCP call — you cannot communicate with the user
or call other tools while it runs. Either samples or time (ms) is required.

**Timing constraint:** The server enforces a 30-second command timeout.
Device initialization adds overhead, so keep capture time under 20 seconds
to avoid "context deadline exceeded" errors. For longer acquisitions, use
multiple shorter captures.

**output_file:** Use a flat filename only — no path separators (/).
Files are saved to the server's working directory.
Good: "my_capture.sr" — Bad: "captures/my_capture.sr"

**Triggers:** 0/1 (level), r/f (rising/falling), e (either edge).
Example: "D0=r" with wait_trigger=true.

## Protocol Decoding

Before decoding, call show_decoder_details to find required channels and options.

**Channel mappings are mandatory.** Without them, decoding produces empty output
with no error. Always include channel assignments in protocol_decoders:
- UART: "uart:rx=D0:baudrate=9600"
- SPI: "spi:clk=D0:mosi=D1:miso=D2:cs=D3"
- I2C: "i2c:scl=D0:sda=D1"

**Decoder options matter.** Some decoders require specific option values to
produce output. For example, the SPI decoder's cs_polarity must match the
target device — set "cs_polarity=active-low" or "active-high" accordingly,
or decoded data will be empty.

## Common Workflows

**Continuous signal** (device is already transmitting, e.g. UART idle traffic):
  Call capture_data with a sufficient time window, then decode.

**User-triggered signal** (user presses a button, plugs a cable, etc.):
  Tell the user the plan and agree on timing BEFORE calling capture_data.
  Example: "I will start a 15-second capture. Please trigger the signal within
  that window." Then capture, then decode.

**Bench instrument measurement** (multimeter, power supply):
  Use serial_query directly — no capture or decoding needed.
  Call get_device_profile first to get correct connection settings.

## Serial Instruments

Call get_device_profile before serial_query — some instruments use non-standard
commands where standard SCPI (e.g. MEAS:VOLT:DC?) fails silently.
For unknown devices, send *IDN? first, then match with get_device_profile.`
