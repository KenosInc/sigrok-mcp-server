# get_device_profile

Look up a device profile by name, model, manufacturer, or `*IDN?` response string. Returns connection settings, supported commands with examples, and device-specific notes.

## Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `query` | string | Yes | Device name, model, manufacturer, or `*IDN?` response string to match against |

## Returns

Device profile including:

- Connection settings (baudrate, parity, etc.)
- Supported commands with examples
- Device-specific notes and limitations

## Examples

### Look up by model name

```
Tool: get_device_profile
Parameters:
  query: XDM1241
```

### Look up by IDN response

```
Tool: get_device_profile
Parameters:
  query: "OWON,XDM1241,24412417,V4.3.0,3"
```

## Notes

- Use this tool before [`serial_query`](serial-query.md) to get the correct settings for a device.
- Profiles are embedded in the server; see [devices](../devices/index.md) for currently supported profiles.
