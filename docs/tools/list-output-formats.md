# list_output_formats

List all supported output file formats.

## Parameters

None.

## Returns

An array of `{id, description}` objects, one per supported output format.

## Example

```
Tool: list_output_formats
```

Response (excerpt):

```json
[
  {"id": "srzip", "description": "sigrok session file format data"},
  {"id": "csv", "description": "Comma-separated values"},
  {"id": "vcd", "description": "Value Change Dump data"}
]
```

## Notes

- Output formats determine the file format used by [`capture_data`](capture-data.md).
