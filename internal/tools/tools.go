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
		mcp.WithDescription("Scan for connected hardware devices. Returns {devices, warnings, hint} where devices is an array of {driver, description} objects. Warnings indicate firmware-related issues for devices that could not be initialized."),
	), h.HandleScanDevices)

	srv.AddTool(mcp.NewTool("check_firmware_status",
		mcp.WithDescription("Check firmware file availability in standard sigrok firmware directories. Returns which directories exist and what firmware files are present. Use this to diagnose device detection issues caused by missing firmware."),
	), h.HandleCheckFirmwareStatus)

	srv.AddTool(mcp.NewTool("capture_data",
		mcp.WithDescription("Capture communication data from a connected device and save to file. Either 'samples' or 'time' must be specified."),
		mcp.WithString("driver", mcp.Description("Hardware driver ID (e.g. 'fx2lafw', 'demo')"), mcp.Required()),
		mcp.WithString("config", mcp.Description("Device configuration (e.g. 'samplerate=1M')")),
		mcp.WithString("channels", mcp.Description("Channels to use (e.g. 'D0,D1,D2')")),
		mcp.WithNumber("samples", mcp.Description("Number of samples to acquire")),
		mcp.WithNumber("time", mcp.Description("How long to sample in milliseconds")),
		mcp.WithString("triggers", mcp.Description("Trigger configuration (e.g. 'D0=r')")),
		mcp.WithBoolean("wait_trigger", mcp.Description("Wait for trigger before capturing")),
		mcp.WithString("output_file", mcp.Description("Output filename (auto-generated if omitted)")),
	), h.HandleCaptureData)

	srv.AddTool(mcp.NewTool("decode_protocol",
		mcp.WithDescription("Decode protocol data from a captured file using sigrok protocol decoders."),
		mcp.WithString("input_file", mcp.Description("Input filename"), mcp.Required()),
		mcp.WithString("protocol_decoders", mcp.Description("Protocol decoders to apply (e.g. 'uart:baudrate=9600')"), mcp.Required()),
		mcp.WithString("input_format", mcp.Description("Input format (e.g. 'vcd', 'binary')")),
		mcp.WithString("annotations", mcp.Description("Decoder annotation filter (e.g. 'uart=rx-data')")),
		mcp.WithBoolean("show_sample_numbers", mcp.Description("Include sample numbers in output")),
		mcp.WithString("meta_output", mcp.Description("Decoder meta output filter (e.g. 'uart=baud')")),
		mcp.WithBoolean("json_trace", mcp.Description("Output in Google Trace Event JSON format")),
	), h.HandleDecodeProtocol)
}
