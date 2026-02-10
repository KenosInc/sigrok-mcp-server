package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/KenosInc/sigrok-mcp-server/internal/sigrok"
	"github.com/mark3labs/mcp-go/mcp"
)

// validIDRe matches valid sigrok identifier strings (alphanumeric, hyphens, underscores).
var validIDRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// validOptionRe matches sigrok-cli option values (config, channels, triggers, decoders, annotations, meta_output).
var validOptionRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._:=,/-]*$`)

// validFilenameRe matches safe filenames (no path separators).
var validFilenameRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._-]*$`)

// Runner abstracts sigrok-cli command execution for testing.
type Runner interface {
	Run(ctx context.Context, args ...string) (*sigrok.CommandResult, error)
}

// Handlers holds MCP tool handler functions.
type Handlers struct {
	runner       Runner
	firmwareDirs []string
}

// NewHandlers creates a new Handlers with the given executor and firmware directories.
func NewHandlers(runner Runner, firmwareDirs []string) *Handlers {
	return &Handlers{runner: runner, firmwareDirs: firmwareDirs}
}

// HandleListSupportedHardware returns all supported hardware drivers.
func (h *Handlers) HandleListSupportedHardware(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.runListSection(ctx, "Supported hardware drivers:")
}

// HandleListSupportedDecoders returns all supported protocol decoders.
func (h *Handlers) HandleListSupportedDecoders(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.runListSection(ctx, "Supported protocol decoders:")
}

// HandleListInputFormats returns all supported input formats.
func (h *Handlers) HandleListInputFormats(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.runListSection(ctx, "Supported input formats:")
}

// HandleListOutputFormats returns all supported output formats.
func (h *Handlers) HandleListOutputFormats(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return h.runListSection(ctx, "Supported output formats:")
}

func (h *Handlers) runListSection(ctx context.Context, sectionHeader string) (*mcp.CallToolResult, error) {
	result, err := h.runner.Run(ctx, "-L")
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		return toolError(result.Stderr), nil
	}

	items := sigrok.ParseListSection(result.Stdout, sectionHeader)
	return jsonResult(items)
}

// HandleShowDecoderDetails returns details for a specific protocol decoder.
func (h *Handlers) HandleShowDecoderDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	decoder := req.GetString("decoder", "")
	if decoder == "" {
		return toolError("missing required parameter: decoder"), nil
	}
	if !validIDRe.MatchString(decoder) {
		return toolError("invalid decoder: must contain only alphanumeric characters, hyphens, and underscores"), nil
	}

	result, err := h.runner.Run(ctx, "--show", "-P", decoder)
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		return toolError(result.Stderr), nil
	}

	details := sigrok.ParseDecoderDetails(result.Stdout)
	return jsonResult(details)
}

// HandleShowDriverDetails returns details for a specific hardware driver.
func (h *Handlers) HandleShowDriverDetails(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driver := req.GetString("driver", "")
	if driver == "" {
		return toolError("missing required parameter: driver"), nil
	}
	if !validIDRe.MatchString(driver) {
		return toolError("invalid driver: must contain only alphanumeric characters, hyphens, and underscores"), nil
	}

	result, err := h.runner.Run(ctx, "--show", "-d", driver)
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		return toolError(result.Stderr), nil
	}

	details := sigrok.ParseDriverDetails(result.Stdout)
	return jsonResult(details)
}

// HandleShowVersion returns sigrok-cli version information.
func (h *Handlers) HandleShowVersion(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := h.runner.Run(ctx, "--version")
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		return toolError(result.Stderr), nil
	}

	info := sigrok.ParseVersion(result.Stdout)
	return jsonResult(info)
}

// HandleScanDevices scans for connected hardware devices.
func (h *Handlers) HandleScanDevices(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := h.runner.Run(ctx, "--scan")
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		msg := nonEmptyError(result)
		if isFirmwareError(msg) {
			msg += "\n\nHint: Some devices require firmware files that are not bundled with the server. " +
				"Use the check_firmware_status tool to diagnose, or mount firmware into the container with: " +
				"docker run -v /path/to/firmware:/usr/local/share/sigrok-firmware:ro"
		}
		return toolError(msg), nil
	}

	scanResult := sigrok.ScanResult{
		Devices: sigrok.ParseScanDevices(result.Stdout),
	}
	if warnings := extractFirmwareWarnings(result.Stderr); len(warnings) > 0 {
		scanResult.Warnings = warnings
		scanResult.Hint = "Some devices may not have been detected due to missing firmware. Use the check_firmware_status tool to diagnose."
	}
	return jsonResult(scanResult)
}

// HandleCaptureData captures communication data from a device and saves to file.
func (h *Handlers) HandleCaptureData(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	driver := req.GetString("driver", "")
	if driver == "" {
		return toolError("missing required parameter: driver"), nil
	}
	if !validIDRe.MatchString(driver) {
		return toolError("invalid driver: must contain only alphanumeric characters, hyphens, and underscores"), nil
	}

	samples := req.GetFloat("samples", 0)
	timeMs := req.GetFloat("time", 0)
	if samples < 0 {
		return toolError("samples must be a positive number"), nil
	}
	if timeMs < 0 {
		return toolError("time must be a positive number"), nil
	}
	if samples <= 0 && timeMs <= 0 {
		return toolError("either 'samples' or 'time' must be specified"), nil
	}
	const maxNumericValue = 1e15
	if samples > maxNumericValue {
		return toolError("samples value is too large"), nil
	}
	if timeMs > maxNumericValue {
		return toolError("time value is too large"), nil
	}

	config := req.GetString("config", "")
	if config != "" && !validOptionRe.MatchString(config) {
		return toolError("invalid config: must contain only alphanumeric characters, dots, underscores, colons, equals, commas, slashes, and hyphens"), nil
	}

	channels := req.GetString("channels", "")
	if channels != "" && !validOptionRe.MatchString(channels) {
		return toolError("invalid channels: must contain only alphanumeric characters, dots, underscores, colons, equals, commas, slashes, and hyphens"), nil
	}

	triggers := req.GetString("triggers", "")
	if triggers != "" && !validOptionRe.MatchString(triggers) {
		return toolError("invalid triggers: must contain only alphanumeric characters, dots, underscores, colons, equals, commas, slashes, and hyphens"), nil
	}

	waitTrigger := req.GetBool("wait_trigger", false)

	outputFile := req.GetString("output_file", "")
	if outputFile != "" && !validFilenameRe.MatchString(outputFile) {
		return toolError("invalid output_file: must contain only alphanumeric characters, dots, underscores, and hyphens (no path separators)"), nil
	}
	if outputFile == "" {
		outputFile = "capture_" + time.Now().UTC().Format("20060102_150405") + ".sr"
	}

	args := []string{"-d", driver}
	if config != "" {
		args = append(args, "-c", config)
	}
	if channels != "" {
		args = append(args, "-C", channels)
	}
	if samples > 0 {
		args = append(args, "--samples", fmt.Sprintf("%d", int64(samples)))
	}
	if timeMs > 0 {
		args = append(args, "--time", fmt.Sprintf("%d", int64(timeMs)))
	}
	if triggers != "" {
		args = append(args, "-t", triggers)
	}
	if waitTrigger {
		args = append(args, "-w")
	}
	args = append(args, "-o", outputFile)

	result, err := h.runner.Run(ctx, args...)
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		return toolError(nonEmptyError(result)), nil
	}

	return jsonResult(sigrok.CaptureResult{
		File:      outputFile,
		RawOutput: result.Stdout,
	})
}

// HandleDecodeProtocol decodes protocol data from a captured file.
func (h *Handlers) HandleDecodeProtocol(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	inputFile := req.GetString("input_file", "")
	if inputFile == "" {
		return toolError("missing required parameter: input_file"), nil
	}
	if !validFilenameRe.MatchString(inputFile) {
		return toolError("invalid input_file: must contain only alphanumeric characters, dots, underscores, and hyphens (no path separators)"), nil
	}

	decoders := req.GetString("protocol_decoders", "")
	if decoders == "" {
		return toolError("missing required parameter: protocol_decoders"), nil
	}
	if !validOptionRe.MatchString(decoders) {
		return toolError("invalid protocol_decoders: must contain only alphanumeric characters, dots, underscores, colons, equals, commas, slashes, and hyphens"), nil
	}

	inputFormat := req.GetString("input_format", "")
	if inputFormat != "" && !validIDRe.MatchString(inputFormat) {
		return toolError("invalid input_format: must contain only alphanumeric characters, hyphens, and underscores"), nil
	}

	annotations := req.GetString("annotations", "")
	if annotations != "" && !validOptionRe.MatchString(annotations) {
		return toolError("invalid annotations: must contain only alphanumeric characters, dots, underscores, colons, equals, commas, slashes, and hyphens"), nil
	}

	showSampleNumbers := req.GetBool("show_sample_numbers", false)

	metaOutput := req.GetString("meta_output", "")
	if metaOutput != "" && !validOptionRe.MatchString(metaOutput) {
		return toolError("invalid meta_output: must contain only alphanumeric characters, dots, underscores, colons, equals, commas, slashes, and hyphens"), nil
	}

	jsonTrace := req.GetBool("json_trace", false)

	args := []string{"-i", inputFile}
	if inputFormat != "" {
		args = append(args, "-I", inputFormat)
	}
	args = append(args, "-P", decoders)
	if annotations != "" {
		args = append(args, "-A", annotations)
	}
	if showSampleNumbers {
		args = append(args, "--protocol-decoder-samplenum")
	}
	if metaOutput != "" {
		args = append(args, "-M", metaOutput)
	}
	if jsonTrace {
		args = append(args, "--protocol-decoder-jsontrace")
	}

	result, err := h.runner.Run(ctx, args...)
	if err != nil {
		return toolError(fmt.Sprintf("sigrok-cli execution failed: %v", err)), nil
	}
	if result.ExitCode != 0 {
		return toolError(nonEmptyError(result)), nil
	}

	format := "text"
	if jsonTrace {
		format = "json_trace"
	}

	return jsonResult(sigrok.DecodeResult{
		Output: result.Stdout,
		Format: format,
	})
}

// HandleCheckFirmwareStatus checks firmware file availability in standard sigrok directories.
func (h *Handlers) HandleCheckFirmwareStatus(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	status := sigrok.FirmwareStatus{
		Directories: make([]sigrok.FirmwareDirectory, 0, len(h.firmwareDirs)),
	}

	for _, dir := range h.firmwareDirs {
		fd := sigrok.FirmwareDirectory{Path: dir}
		entries, err := os.ReadDir(dir)
		if err != nil {
			fd.Exists = false
			status.Directories = append(status.Directories, fd)
			continue
		}
		fd.Exists = true
		for _, e := range entries {
			if !e.IsDir() {
				fd.Files = append(fd.Files, e.Name())
			}
		}
		status.TotalFiles += len(fd.Files)
		status.Directories = append(status.Directories, fd)
	}

	if status.TotalFiles == 0 {
		status.Hint = "No firmware files found. Some hardware drivers (e.g. kingst-la2016, saleae-logic16) " +
			"require firmware files that cannot be redistributed. Mount your firmware directory into the container: " +
			"docker run -v /path/to/firmware:/usr/local/share/sigrok-firmware:ro ..."
	}

	return jsonResult(status)
}

// isFirmwareError checks if an error message indicates a firmware-related failure.
func isFirmwareError(msg string) bool {
	lower := strings.ToLower(msg)
	return strings.Contains(lower, "firmware") ||
		strings.Contains(lower, "failed to open resource") ||
		strings.Contains(lower, ".fw") ||
		strings.Contains(lower, ".bitstream")
}

// extractFirmwareWarnings extracts firmware-related warning lines from stderr.
func extractFirmwareWarnings(stderr string) []string {
	if stderr == "" {
		return nil
	}
	var warnings []string
	for _, line := range strings.Split(stderr, "\n") {
		if isFirmwareError(line) {
			trimmed := strings.TrimSpace(line)
			if trimmed != "" {
				warnings = append(warnings, trimmed)
			}
		}
	}
	return warnings
}

func nonEmptyError(result *sigrok.CommandResult) string {
	if result.Stderr != "" {
		return result.Stderr
	}
	msg := fmt.Sprintf("sigrok-cli exited with code %d", result.ExitCode)
	if result.Stdout != "" {
		msg += ": " + result.Stdout
	}
	return msg
}

func jsonResult(v any) (*mcp.CallToolResult, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal result: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: string(data)},
		},
	}, nil
}

func toolError(msg string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{Type: "text", Text: msg},
		},
		IsError: true,
	}
}
