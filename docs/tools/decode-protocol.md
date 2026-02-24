# decode_protocol

Decode protocol data from a captured file using sigrok protocol decoders.

## Parameters

| Parameter | Type | Required | Description |
|---|---|---|---|
| `input_file` | string | Yes | Input filename |
| `protocol_decoders` | string | Yes | Protocol decoders to apply (e.g. `uart:baudrate=9600`) |
| `input_format` | string | No | Input format (e.g. `vcd`, `binary`) |
| `annotations` | string | No | Decoder annotation filter (e.g. `uart=rx-data`) |
| `show_sample_numbers` | boolean | No | Include sample numbers in output |
| `meta_output` | string | No | Decoder meta output filter (e.g. `uart=baud`) |
| `json_trace` | boolean | No | Output in Google Trace Event JSON format |

## Returns

Decoded protocol data from the file.

## Examples

### Decode UART data

```
Tool: decode_protocol
Parameters:
  input_file: capture.sr
  protocol_decoders: uart:baudrate=9600:rx=D0
```

### Decode SPI with annotation filter

```
Tool: decode_protocol
Parameters:
  input_file: capture.sr
  protocol_decoders: spi:clk=D0:mosi=D1:miso=D2:cs=D3
  annotations: spi=mosi-data
```

### Decode with JSON trace output

```
Tool: decode_protocol
Parameters:
  input_file: capture.sr
  protocol_decoders: i2c:scl=D0:sda=D1
  json_trace: true
```

## Notes

- Use [`show_decoder_details`](show-decoder-details.md) to see available options and channels for a decoder.
- Decoder options are specified inline with the decoder ID, separated by colons (e.g. `uart:baudrate=115200:rx=D0`).
- Files are captured using [`capture_data`](capture-data.md) or imported from external sources.
