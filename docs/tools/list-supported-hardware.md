# list_supported_hardware

List all supported hardware drivers.

## Parameters

None.

## Returns

An array of `{id, description}` objects, one per supported hardware driver.

## Example

```
Tool: list_supported_hardware
```

Response (excerpt):

```json
[
  {"id": "demo", "description": "Demo driver and target"},
  {"id": "fx2lafw", "description": "fx2lafw (generic driver for FX2 based LAs)"},
  {"id": "kingst-la2016", "description": "Kingst LA2016"}
]
```

## Notes

- This is a read-only query that does not require connected hardware.
- The list comes from the sigrok-cli installation inside the container.
