# Quick Start

After [installing](installation.md) the server, configure your MCP client to connect to it.

## Claude Desktop

Add to your `claude_desktop_config.json`:

=== "Basic"

    ```json
    {
      "mcpServers": {
        "sigrok": {
          "command": "docker",
          "args": ["run", "-i", "--rm", "sigrok-mcp-server"]
        }
      }
    }
    ```

=== "With USB access"

    ```json
    {
      "mcpServers": {
        "sigrok": {
          "command": "docker",
          "args": ["run", "-i", "--rm", "--privileged", "sigrok-mcp-server"]
        }
      }
    }
    ```

=== "With USB + firmware"

    ```json
    {
      "mcpServers": {
        "sigrok": {
          "command": "docker",
          "args": [
            "run", "-i", "--rm", "--privileged",
            "-v", "/path/to/sigrok-firmware:/usr/local/share/sigrok-firmware:ro",
            "sigrok-mcp-server"
          ]
        }
      }
    }
    ```

## Claude Code

=== "Basic"

    ```bash
    claude mcp add sigrok -- docker run -i --rm sigrok-mcp-server
    ```

=== "With USB + firmware"

    ```bash
    claude mcp add sigrok -- docker run -i --rm --privileged \
      -v /path/to/sigrok-firmware:/usr/local/share/sigrok-firmware:ro \
      sigrok-mcp-server
    ```

## Verify the connection

Once configured, ask the LLM to run the `show_version` tool. You should see sigrok-cli and library version information returned, confirming the server is running correctly.

## What's next?

- Browse [available tools](../tools/index.md) to see what you can do
- Check the [workflow example](../guides/workflow-example.md) for a typical signal analysis session
- If your device needs firmware, see the [firmware guide](../getting-started/firmware.md)
