package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/KenosInc/sigrok-mcp-server/internal/sigrok"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// mockExecutor implements a test double for sigrok.Executor.
type mockExecutor struct {
	result  *sigrok.CommandResult
	err     error
	gotArgs []string
}

func (m *mockExecutor) Run(_ context.Context, args ...string) (*sigrok.CommandResult, error) {
	m.gotArgs = args
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

func TestRegisterAll(t *testing.T) {
	srv := server.NewMCPServer("test", "0.0.1")
	h := NewHandlers(&mockExecutor{})
	RegisterAll(srv, h)

	ctx := context.Background()

	// Initialize the server first (required before tools/list).
	initMsg := json.RawMessage(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}`)
	resp := srv.HandleMessage(ctx, initMsg)
	if resp == nil {
		t.Fatal("expected initialize response")
	}

	// Send tools/list request.
	listMsg := json.RawMessage(`{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}`)
	resp = srv.HandleMessage(ctx, listMsg)
	if resp == nil {
		t.Fatal("expected tools/list response")
	}

	// Parse the response to extract tool names.
	respBytes, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var parsed struct {
		Result struct {
			Tools []struct {
				Name        string `json:"name"`
				InputSchema struct {
					Required []string `json:"required"`
				} `json:"inputSchema"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(respBytes, &parsed); err != nil {
		t.Fatalf("failed to parse tools/list response: %v", err)
	}

	wantTools := map[string]bool{
		"list_input_formats":     true,
		"list_output_formats":    true,
		"list_supported_decoders": true,
		"list_supported_hardware": true,
		"scan_devices":           true,
		"show_decoder_details":   true,
		"show_driver_details":    true,
		"show_version":           true,
	}

	if len(parsed.Result.Tools) != len(wantTools) {
		t.Fatalf("got %d tools, want %d", len(parsed.Result.Tools), len(wantTools))
	}

	for _, tool := range parsed.Result.Tools {
		if !wantTools[tool.Name] {
			t.Errorf("unexpected tool: %q", tool.Name)
		}
		// Verify parameterized tools have required params.
		switch tool.Name {
		case "show_decoder_details":
			if !contains(tool.InputSchema.Required, "decoder") {
				t.Errorf("show_decoder_details missing required param 'decoder'")
			}
		case "show_driver_details":
			if !contains(tool.InputSchema.Required, "driver") {
				t.Errorf("show_driver_details missing required param 'driver'")
			}
		}
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func TestHandleShowVersion(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "sigrok-cli 0.7.2\n\nLibraries and features:\n- libsigrok 0.5.2\n- libsigrokdecode 0.5.3\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleShowVersion(context.Background(), makeRequest("show_version", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"--version"}) {
		t.Errorf("args = %v, want [--version]", mock.gotArgs)
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
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported hardware drivers:\n  demo                 Demo driver and pattern generator\n  fx2lafw              fx2lafw\n\nSupported input formats:\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleListSupportedHardware(context.Background(), makeRequest("list_supported_hardware", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"-L"}) {
		t.Errorf("args = %v, want [-L]", mock.gotArgs)
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
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported protocol decoders:\n  uart                 UART\n  spi                  SPI\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleListSupportedDecoders(context.Background(), makeRequest("list_supported_decoders", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"-L"}) {
		t.Errorf("args = %v, want [-L]", mock.gotArgs)
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
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported input formats:\n  csv                  Comma-separated values\n\nSupported output formats:\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleListInputFormats(context.Background(), makeRequest("list_input_formats", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"-L"}) {
		t.Errorf("args = %v, want [-L]", mock.gotArgs)
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
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Supported output formats:\n  csv                  Comma-separated values\n  vcd                  Value Change Dump data\n\nSupported transform modules:\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleListOutputFormats(context.Background(), makeRequest("list_output_formats", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"-L"}) {
		t.Errorf("args = %v, want [-L]", mock.gotArgs)
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
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "ID: uart\nName: UART\nLong name: Universal Asynchronous Receiver/Transmitter\nDescription: Asynchronous, serial bus.\nLicense: gplv2+\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", map[string]any{"decoder": "uart"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"--show", "-P", "uart"}) {
		t.Errorf("args = %v, want [--show -P uart]", mock.gotArgs)
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

func TestHandleShowDecoderDetailsInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{})

	tests := []struct {
		name    string
		decoder string
	}{
		{"flag injection", "--output-file=/tmp/evil"},
		{"spaces", "uart spi"},
		{"special chars", "uart;rm -rf"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", map[string]any{"decoder": tt.decoder}))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleShowDriverDetails(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "Driver functions:\n    Demo device\nScan options:\n    logic_channels\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleShowDriverDetails(context.Background(), makeRequest("show_driver_details", map[string]any{"driver": "demo"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"--show", "-d", "demo"}) {
		t.Errorf("args = %v, want [--show -d demo]", mock.gotArgs)
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

func TestHandleShowDriverDetailsInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{})

	result, err := h.HandleShowDriverDetails(context.Background(), makeRequest("show_driver_details", map[string]any{"driver": "--evil-flag"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleScanDevices(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "The following devices were found:\ndemo - Demo device with 13 channels: D0 D1 D2 D3 D4 D5 D6 D7 A0 A1 A2 A3 A4\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"--scan"}) {
		t.Errorf("args = %v, want [--scan]", mock.gotArgs)
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

	// Execution errors should be returned as tool errors (IsError=true),
	// not as Go errors, so LLMs can see the failure message.
	result, err := h.HandleShowVersion(context.Background(), makeRequest("show_version", nil))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text == "" {
		t.Error("expected non-empty error message")
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
