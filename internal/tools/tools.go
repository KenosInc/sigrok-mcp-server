package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterAll registers all sigrok MCP tools on the given server.
func RegisterAll(srv *server.MCPServer, h *Handlers) {
	srv.AddTool(mcp.NewTool("list_supported_hardware",
		mcp.WithDescription("List all supported hardware drivers. Returns an array of {id, description} objects."),
	), h.HandleListSupportedHardware)

	srv.AddTool(mcp.NewTool("list_supported_decoders",
		mcp.WithDescription("List all supported protocol decoders. Returns an array of {id, description} objects."),
	), h.HandleListSupportedDecoders)

	srv.AddTool(mcp.NewTool("list_input_formats",
		mcp.WithDescription("List all supported input file formats. Returns an array of {id, description} objects."),
	), h.HandleListInputFormats)

	srv.AddTool(mcp.NewTool("list_output_formats",
		mcp.WithDescription("List all supported output file formats. Returns an array of {id, description} objects."),
	), h.HandleListOutputFormats)

	srv.AddTool(mcp.NewTool("show_decoder_details",
		mcp.WithDescription("Show detailed information about a specific protocol decoder, including options, channels, annotation classes, and documentation."),
		mcp.WithString("decoder",
			mcp.Description("Protocol decoder ID (e.g. 'uart', 'spi', 'i2c')"),
			mcp.Required(),
		),
	), h.HandleShowDecoderDetails)

	srv.AddTool(mcp.NewTool("show_driver_details",
		mcp.WithDescription("Show detailed information about a specific hardware driver, including supported functions, scan options, and connected devices."),
		mcp.WithString("driver",
			mcp.Description("Hardware driver ID (e.g. 'demo', 'fx2lafw', 'rigol-ds')"),
			mcp.Required(),
		),
	), h.HandleShowDriverDetails)

	srv.AddTool(mcp.NewTool("show_version",
		mcp.WithDescription("Show sigrok-cli version information, including library versions."),
	), h.HandleShowVersion)

	srv.AddTool(mcp.NewTool("scan_devices",
		mcp.WithDescription("Scan for connected hardware devices. Returns an array of {driver, description} objects."),
	), h.HandleScanDevices)
}
