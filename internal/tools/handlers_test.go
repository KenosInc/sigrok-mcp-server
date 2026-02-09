package tools

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/KenosInc/sigrok-mcp-server/internal/sigrok"
	"github.com/mark3labs/mcp-go/mcp"
)

// mockExecutor implements a test double for sigrok.Executor.
type mockExecutor struct {
	result *sigrok.CommandResult
	err    error
}

func (m *mockExecutor) Run(_ context.Context, _ ...string) (*sigrok.CommandResult, error) {
	return m.result, m.err
}

func makeRequest(name string, args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

func assertTextResult(t *testing.T, result *mcp.CallToolResult, errExpected bool) string {
	t.Helper()
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.IsError != errExpected {
		t.Errorf("IsError = %v, want %v", result.IsError, errExpected)
	}
	if len(result.Content) == 0 {
		t.Fatal("expected at least one content item")
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		t.Fatalf("expected TextContent, got %T", result.Content[0])
	}
	return tc.Text
}

func TestHandleShowVersion(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "sigrok-cli 0.7.2\n\nLibraries and features:\n- libsigrok 0.5.2\n- libsigrokdecode 0.5.3\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleShowVersion(context.Background(), makeRequest("show_version", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.VersionInfo
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON result: %v", err)
	}
	if parsed.CLIVersion != "0.7.2" {
		t.Errorf("CLIVersion = %q, want %q", parsed.CLIVersion, "0.7.2")
	}
}

func TestHandleListSupportedHardware(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported hardware drivers:\n  demo                 Demo driver and pattern generator\n  fx2lafw              fx2lafw\n\nSupported input formats:\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleListSupportedHardware(context.Background(), makeRequest("list_supported_hardware", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var items []sigrok.ListItem
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("failed to parse JSON result: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if items[0].ID != "demo" {
		t.Errorf("first item ID = %q, want %q", items[0].ID, "demo")
	}
}

func TestHandleListSupportedDecoders(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported protocol decoders:\n  uart                 UART\n  spi                  SPI\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleListSupportedDecoders(context.Background(), makeRequest("list_supported_decoders", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var items []sigrok.ListItem
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestHandleListInputFormats(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported input formats:\n  csv                  Comma-separated values\n\nSupported output formats:\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleListInputFormats(context.Background(), makeRequest("list_input_formats", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var items []sigrok.ListItem
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(items) != 1 || items[0].ID != "csv" {
		t.Errorf("unexpected items: %v", items)
	}
}

func TestHandleListOutputFormats(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported output formats:\n  csv                  Comma-separated values\n  vcd                  Value Change Dump data\n\nSupported transform modules:\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleListOutputFormats(context.Background(), makeRequest("list_output_formats", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var items []sigrok.ListItem
	if err := json.Unmarshal([]byte(text), &items); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestHandleShowDecoderDetails(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "ID: uart\nName: UART\nLong name: Universal Asynchronous Receiver/Transmitter\nDescription: Asynchronous, serial bus.\nLicense: gplv2+\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", map[string]any{"decoder": "uart"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.DecoderDetails
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.ID != "uart" {
		t.Errorf("ID = %q, want %q", parsed.ID, "uart")
	}
}

func TestHandleShowDecoderDetailsMissingParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{})

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleShowDriverDetails(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Driver functions:\n    Demo device\nScan options:\n    logic_channels\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleShowDriverDetails(context.Background(), makeRequest("show_driver_details", map[string]any{"driver": "demo"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.DriverDetails
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(parsed.Functions) == 0 {
		t.Error("expected at least one function")
	}
}

func TestHandleShowDriverDetailsMissingParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{})

	result, err := h.HandleShowDriverDetails(context.Background(), makeRequest("show_driver_details", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleScanDevices(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "The following devices were found:\ndemo - Demo device with 13 channels: D0 D1 D2 D3 D4 D5 D6 D7 A0 A1 A2 A3 A4\n",
			ExitCode: 0,
		},
	})

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var devices []sigrok.ScannedDevice
	if err := json.Unmarshal([]byte(text), &devices); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Driver != "demo" {
		t.Errorf("driver = %q, want %q", devices[0].Driver, "demo")
	}
}

func TestHandlerExecutionError(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		err: errors.New("binary not found"),
	})

	result, err := h.HandleShowVersion(context.Background(), makeRequest("show_version", nil))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestHandlerNonZeroExit(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			Stderr:   "Error: unknown protocol decoder 'foo'.\n",
			ExitCode: 1,
		},
	})

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", map[string]any{"decoder": "foo"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}
