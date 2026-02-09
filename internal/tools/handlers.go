package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/KenosInc/sigrok-mcp-server/internal/sigrok"
	"github.com/mark3labs/mcp-go/mcp"
)

// validIDRe matches valid sigrok identifier strings (alphanumeric, hyphens, underscores).
var validIDRe = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

// Runner abstracts sigrok-cli command execution for testing.
type Runner interface {
	Run(ctx context.Context, args ...string) (*sigrok.CommandResult, error)
}

// Handlers holds MCP tool handler functions.
type Handlers struct {
	runner Runner
}

// NewHandlers creates a new Handlers with the given executor.
func NewHandlers(runner Runner) *Handlers {
	return &Handlers{runner: runner}
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
		return toolError(result.Stderr), nil
	}

	devices := sigrok.ParseScanDevices(result.Stdout)
	return jsonResult(devices)
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
