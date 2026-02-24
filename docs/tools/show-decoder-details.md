# show_decoder_details

Show detailed information about a specific protocol decoder, including options, channels, annotation classes, and documentation.

## Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `decoder` | string | Yes | Protocol decoder ID (e.g. `uart`, `spi`, `i2c`) |

## Returns

Detailed information about the decoder including:

- Description and documentation
- Required and optional channels
- Configurable options with defaults
- Annotation classes (output categories)

## Example

```
Tool: show_decoder_details
Parameters:
  decoder: uart
```

## Notes

- Use [`list_supported_decoders`](list-supported-decoders.md) to discover available decoder IDs.
- Decoder options (e.g. `baudrate`, `parity`) are passed via the `protocol_decoders` parameter in [`decode_protocol`](decode-protocol.md).
