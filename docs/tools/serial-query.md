# serial_query

Send a command string over a serial port and return the device response. Works with any serial-attached instrument that accepts text commands (e.g. SCPI). Independent of sigrok-cli.

## Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `port` | string | Yes | Serial device path (e.g. `/dev/ttyUSB0`) |
| `command` | string | Yes | Command to send (e.g. `*IDN?`, `MEAS:VOLT:DC?`) |
| `baudrate` | number | No | Baud rate (default: 9600) |
| `databits` | number | No | Data bits: 5, 6, 7, or 8 (default: 8) |
| `parity` | string | No | Parity: `none`, `odd`, `even`, `mark`, `space` (default: `none`) |
| `stopbits` | string | No | Stop bits: `1`, `1.5`, `2` (default: `1`) |
| `timeout_ms` | number | No | Read timeout in milliseconds (default: 1000) |

## Returns

The device response string.

## Examples

### Identify a device

```
Tool: serial_query
Parameters:
  port: /dev/ttyUSB0
  command: "*IDN?"
  baudrate: 115200
  timeout_ms: 3000
```

### Read a voltage measurement

```
Tool: serial_query
Parameters:
  port: /dev/ttyUSB0
  command: "MEAS:VOLT:DC?"
  baudrate: 9600
```

## Notes

- This tool communicates directly over serial and does **not** use sigrok-cli.
- Use [`get_device_profile`](get-device-profile.md) to look up correct connection settings before querying an unknown device.
- See [device documentation](../devices/index.md) for device-specific command references.
