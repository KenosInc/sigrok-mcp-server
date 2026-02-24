# list_input_formats

List all supported input file formats.

## Parameters

None.

## Returns

An array of `{id, description}` objects, one per supported input format.

## Example

```
Tool: list_input_formats
```

Response (excerpt):

```json
[
  {"id": "binary", "description": "Raw binary logic data"},
  {"id": "vcd", "description": "Value Change Dump data"},
  {"id": "wav", "description": "WAV file"}
]
```

## Notes

- Input formats determine what file types can be used with [`decode_protocol`](decode-protocol.md).
