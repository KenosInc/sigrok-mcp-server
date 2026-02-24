# Firmware

Some hardware drivers require firmware files that cannot be bundled with this server due to licensing restrictions.

## Which devices need firmware?

Devices that require firmware upload on connection include:

- **Kingst LA2016** / LA1016 / LA5016
- **Saleae Logic16**
- **fx2lafw**-based analyzers (many low-cost USB logic analyzers)

Devices that do **not** need firmware:

- The `demo` virtual driver (always works)
- Protocol-only analysis (decoding from files)
- Some standalone instruments (SCPI multimeters, oscilloscopes)

## Checking firmware status

Use the `check_firmware_status` tool to verify firmware availability:

```
Tool: check_firmware_status
```

This reports which firmware directories exist and what files are present, helping diagnose device detection issues.

## Providing firmware (Docker)

Mount your firmware directory into the container:

```bash
docker run -i --rm --privileged \
  -v /path/to/sigrok-firmware:/usr/local/share/sigrok-firmware:ro \
  sigrok-mcp-server
```

## Where to get firmware

See the [sigrok wiki on firmware](https://sigrok.org/wiki/Firmware) for extraction instructions for your specific device. Each device has different firmware sources and extraction methods.

!!! note
    Firmware files are typically extracted from vendor software installers. The sigrok wiki provides step-by-step instructions for each supported device.
