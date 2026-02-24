# scan_devices

Scan for connected hardware devices.

## Parameters

All parameters are optional. When no parameters are provided, a broad scan is performed.

| Parameter | Type | Required | Description |
|---|---|---|---|
| `driver` | string | No | Hardware driver ID for targeted scanning (e.g. `scpi-dmm`, `rigol-ds`) |
| `conn` | string | No | Connection string for targeted scanning (e.g. `/dev/ttyUSB0:serialcomm=115200/8n1`) |

## Returns

An object with:

- `devices` — Array of `{driver, description}` objects for detected devices
- `warnings` — Firmware-related issues for devices that could not be initialized
- `hint` — Suggestions for resolving issues

## Examples

### Broad scan

```
Tool: scan_devices
```

### Targeted scan for a serial instrument

```
Tool: scan_devices
Parameters:
  driver: scpi-dmm
  conn: /dev/ttyUSB0:serialcomm=115200/8n1
```

## Notes

- Broad scans may take several seconds depending on available drivers.
- Warnings about firmware typically indicate that a device was detected but its firmware files are missing. See the [firmware guide](../getting-started/firmware.md).
- SCPI driver categories: `scpi-dmm` (multimeters), `scpi-pps` (power supplies), or vendor-specific like `rigol-ds` (oscilloscopes).
