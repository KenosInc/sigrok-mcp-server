# Troubleshooting

## Device not found

**Symptom:** `scan_devices` returns no devices.

**Possible causes:**

1. **USB not passed through (Docker):** Use `--privileged` or `--device` flag.

    ```bash
    docker run -i --rm --privileged sigrok-mcp-server
    ```

2. **Missing firmware:** Run `check_firmware_status` to verify. See the [firmware guide](../getting-started/firmware.md).

3. **Permission denied:** On Linux, ensure your user is in the `plugdev` group, or use udev rules for the device.

## Timeout errors

**Symptom:** Commands fail with a timeout error.

**Possible causes:**

1. **Slow device:** Increase the timeout via environment variable:

    ```bash
    docker run -i --rm -e SIGROK_TIMEOUT_SECONDS=120 sigrok-mcp-server
    ```

2. **Serial timeout too short:** For `serial_query`, increase `timeout_ms` (some instruments are slow to respond):

    ```
    Tool: serial_query
    Parameters:
      port: /dev/ttyUSB0
      command: "*IDN?"
      timeout_ms: 5000
    ```

## Decoder produces no output

**Symptom:** `decode_protocol` returns empty results.

**Possible causes:**

1. **Wrong channel mapping:** Ensure decoder channels match the capture channels. Use `show_decoder_details` to see required channels.

2. **Wrong baud rate / protocol settings:** Verify settings match the actual signal.

3. **No transitions in data:** The capture may not contain the expected signal activity.

## sigrok-cli not found

**Symptom:** Server fails to start or tools return "sigrok-cli not found" errors.

**Possible causes:**

1. **Not using Docker:** The Docker image bundles sigrok-cli. If running from source, install it separately.

2. **Custom path:** Set `SIGROK_CLI_PATH` if sigrok-cli is installed in a non-standard location.

## Serial port access denied

**Symptom:** `serial_query` returns a permission error.

**Possible causes:**

1. **Docker:** The serial device must be passed through:

    ```bash
    docker run -i --rm --device=/dev/ttyUSB0 sigrok-mcp-server
    ```

2. **Linux permissions:** Add your user to the `dialout` group:

    ```bash
    sudo usermod -a -G dialout $USER
    ```
