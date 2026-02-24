# list_supported_decoders

List all supported protocol decoders.

## Parameters

None.

## Returns

An array of `{id, description}` objects, one per supported protocol decoder.

## Example

```
Tool: list_supported_decoders
```

Response (excerpt):

```json
[
  {"id": "uart", "description": "Universal Asynchronous Receiver/Transmitter"},
  {"id": "spi", "description": "Serial Peripheral Interface"},
  {"id": "i2c", "description": "Inter-Integrated Circuit"}
]
```

## Notes

- This is a read-only query that does not require connected hardware.
- Use [`show_decoder_details`](show-decoder-details.md) to get full information about a specific decoder.
