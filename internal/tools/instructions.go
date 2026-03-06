package tools

// Instructions is sent to LLM clients as part of the MCP initialize response.
// It provides guidance on how to use this server's tools effectively.
const Instructions = `# sigrok MCP Server — Usage Guide

## Signal Capture

1. Use scan_devices to find hardware. If empty, call check_firmware_status
   — USB analyzers often fail silently without firmware.
2. Use show_driver_details to check supported config options before capturing.
3. capture_data blocks the entire MCP call — you cannot communicate with the user
   or call other tools while it runs. Either samples or time (ms) is required.
4. Before decoding, call show_decoder_details to find required channels.
   You MUST include channel mappings in protocol_decoders — without them,
   decoding produces empty output with no error.
   Examples: "uart:rx=D0:baudrate=9600", "spi:clk=D0:mosi=D1:miso=D2:cs=D3",
   "i2c:scl=D0:sda=D1".

Trigger syntax: 0/1 (level), r/f (rising/falling), e (either edge).
Example: "D0=r" with wait_trigger=true.

### Use cases

**Continuous signal** (device is already transmitting, e.g. UART idle traffic):
  Just call capture_data with a sufficient time window, then decode.

**User-triggered signal** (user presses a button, plugs a cable, etc.):
  Tell the user the plan and agree on timing BEFORE calling capture_data.
  Example: "I will start a 15-second capture. Please trigger the signal within
  that window." Then capture, then decode.

**Read a measurement from a bench instrument** (multimeter, power supply):
  Use serial_query directly — no capture or decoding needed.
  Call get_device_profile first to get correct connection settings.

## Serial Instruments

Call get_device_profile before serial_query — some instruments use non-standard
commands where standard SCPI (e.g. MEAS:VOLT:DC?) fails silently.
For unknown devices, send *IDN? first, then match with get_device_profile.`
