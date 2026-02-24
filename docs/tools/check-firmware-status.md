# check_firmware_status

Check firmware file availability in standard sigrok firmware directories.

## Parameters

None.

## Returns

Information about:

- Which firmware directories exist
- What firmware files are present in each directory
- Overall firmware status

## Example

```
Tool: check_firmware_status
```

## Notes

- Use this tool to diagnose device detection issues — missing firmware is the most common cause.
- See the [firmware guide](../getting-started/firmware.md) for how to provide firmware files.
