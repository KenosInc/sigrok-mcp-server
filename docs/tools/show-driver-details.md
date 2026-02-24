# show_driver_details

Show detailed information about a specific hardware driver, including supported functions, scan options, and connected devices.

## Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `driver` | string | Yes | Hardware driver ID (e.g. `demo`, `fx2lafw`, `rigol-ds`) |

## Returns

Detailed information about the driver including:

- Supported functions and capabilities
- Scan options (connection parameters, serial settings)
- Detected devices (if any are connected)

## Example

```
Tool: show_driver_details
Parameters:
  driver: fx2lafw
```

## Notes

- Use [`list_supported_hardware`](list-supported-hardware.md) to discover available driver IDs.
- This tool may trigger a device scan for the specified driver.
