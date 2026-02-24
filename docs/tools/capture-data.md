# capture_data

Capture communication data from a connected device and save to file.

## Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `driver` | string | Yes | Hardware driver ID (e.g. `fx2lafw`, `demo`) |
| `conn` | string | No | Connection string for serial/network devices |
| `config` | string | No | Device configuration (e.g. `samplerate=1M`) |
| `channels` | string | No | Channels to use (e.g. `D0,D1,D2`) |
| `samples` | number | No* | Number of samples to acquire |
| `time` | number | No* | How long to sample in milliseconds |
| `triggers` | string | No | Trigger configuration (e.g. `D0=r`) |
| `wait_trigger` | boolean | No | Wait for trigger before capturing |
| `output_file` | string | No | Output filename (auto-generated if omitted) |

*Either `samples` or `time` must be specified.

## Returns

Information about the captured file, including the path to the saved data.

## Examples

### Capture with the demo driver

```
Tool: capture_data
Parameters:
  driver: demo
  config: samplerate=1M
  samples: 1000000
```

### Capture from a real device with trigger

```
Tool: capture_data
Parameters:
  driver: fx2lafw
  channels: D0,D1
  config: samplerate=4M
  samples: 2000000
  triggers: D0=r
  wait_trigger: true
```

## Notes

- The output file can be passed to [`decode_protocol`](decode-protocol.md) for analysis.
- If `output_file` is omitted, a filename is generated automatically.
- Use [`scan_devices`](scan-devices.md) first to discover available devices and their drivers.
