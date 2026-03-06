package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/KenosInc/sigrok-mcp-server/internal/devices"
	"github.com/KenosInc/sigrok-mcp-server/internal/serial"
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
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)
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
		"list_input_formats":      true,
		"list_output_formats":     true,
		"list_supported_decoders": true,
		"list_supported_hardware": true,
		"scan_devices":            true,
		"show_decoder_details":    true,
		"show_driver_details":     true,
		"show_version":            true,
		"capture_data":            true,
		"decode_protocol":         true,
		"check_firmware_status":   true,
		"serial_query":            true,
		"get_device_profile":      true,
		"render_waveform":         true,
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
		case "capture_data":
			if !contains(tool.InputSchema.Required, "driver") {
				t.Errorf("capture_data missing required param 'driver'")
			}
		case "decode_protocol":
			if !contains(tool.InputSchema.Required, "input_file") {
				t.Errorf("decode_protocol missing required param 'input_file'")
			}
			if !contains(tool.InputSchema.Required, "protocol_decoders") {
				t.Errorf("decode_protocol missing required param 'protocol_decoders'")
			}
		case "serial_query":
			if !contains(tool.InputSchema.Required, "port") {
				t.Errorf("serial_query missing required param 'port'")
			}
			if !contains(tool.InputSchema.Required, "command") {
				t.Errorf("serial_query missing required param 'command'")
			}
		case "get_device_profile":
			if !contains(tool.InputSchema.Required, "query") {
				t.Errorf("get_device_profile missing required param 'query'")
			}
		case "render_waveform":
			if !contains(tool.InputSchema.Required, "input_file") {
				t.Errorf("render_waveform missing required param 'input_file'")
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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleShowDecoderDetailsInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

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
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleShowDriverDetails(context.Background(), makeRequest("show_driver_details", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleShowDriverDetailsInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

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
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"--scan"}) {
		t.Errorf("args = %v, want [--scan]", mock.gotArgs)
	}

	text := assertTextResult(t, result, false)
	var scanResult sigrok.ScanResult
	if err := json.Unmarshal([]byte(text), &scanResult); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(scanResult.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(scanResult.Devices))
	}
	if scanResult.Devices[0].Driver != "demo" {
		t.Errorf("driver = %q, want %q", scanResult.Devices[0].Driver, "demo")
	}
	if len(scanResult.Warnings) != 0 {
		t.Errorf("expected no warnings, got %v", scanResult.Warnings)
	}
}

func TestHandleScanDevicesWithFirmwareWarnings(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "The following devices were found:\ndemo - Demo device with 13 channels\n",
			Stderr:   "sr: kingst-la2016: Failed to open resource 'kingst-la-01a2.fw'\nsr: kingst-la2016: MCU firmware upload failed\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var scanResult sigrok.ScanResult
	if err := json.Unmarshal([]byte(text), &scanResult); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(scanResult.Warnings) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %v", len(scanResult.Warnings), scanResult.Warnings)
	}
	if scanResult.Hint == "" {
		t.Error("expected non-empty hint")
	}
}

func TestHandleScanDevicesFirmwareError(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stderr:   "Failed to open resource 'kingst-la-01a2.fw'",
			ExitCode: 1,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "Hint:") {
		t.Errorf("expected firmware hint in error, got %q", text)
	}
	if !strings.Contains(text, "check_firmware_status") {
		t.Errorf("expected check_firmware_status reference in error, got %q", text)
	}
}

func TestHandleCheckFirmwareStatusWithFiles(t *testing.T) {
	tmpDir := t.TempDir()
	for _, name := range []string{"kingst-la-01a2.fw", "kingst-la2016-fpga.bitstream"} {
		if err := os.WriteFile(tmpDir+"/"+name, []byte("test"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	h := NewHandlers(&mockExecutor{}, []string{tmpDir, "/nonexistent/path"}, nil, nil)

	result, err := h.HandleCheckFirmwareStatus(context.Background(), makeRequest("check_firmware_status", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var status sigrok.FirmwareStatus
	if err := json.Unmarshal([]byte(text), &status); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if status.TotalFiles != 2 {
		t.Errorf("total_files = %d, want 2", status.TotalFiles)
	}
	if len(status.Directories) != 2 {
		t.Fatalf("expected 2 directories, got %d", len(status.Directories))
	}
	if !status.Directories[0].Exists {
		t.Error("expected first directory to exist")
	}
	if status.Directories[1].Exists {
		t.Error("expected second directory to not exist")
	}
	if status.Hint != "" {
		t.Errorf("expected empty hint when files are present, got %q", status.Hint)
	}
}

func TestHandleCheckFirmwareStatusEmpty(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, []string{"/nonexistent/path1", "/nonexistent/path2"}, nil, nil)

	result, err := h.HandleCheckFirmwareStatus(context.Background(), makeRequest("check_firmware_status", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var status sigrok.FirmwareStatus
	if err := json.Unmarshal([]byte(text), &status); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if status.TotalFiles != 0 {
		t.Errorf("total_files = %d, want 0", status.TotalFiles)
	}
	if status.Hint == "" {
		t.Error("expected non-empty hint when no firmware found")
	}
}

func TestHandlerExecutionError(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		err: errors.New("binary not found"),
	}, nil, nil, nil)

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

func TestHandleCaptureData(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "fx2lafw",
		"samples": float64(10000),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify args: -d fx2lafw --samples 10000 -o <auto>
	if len(mock.gotArgs) < 5 {
		t.Fatalf("expected at least 5 args, got %v", mock.gotArgs)
	}
	if mock.gotArgs[0] != "-d" || mock.gotArgs[1] != "fx2lafw" {
		t.Errorf("expected -d fx2lafw, got %v", mock.gotArgs[:2])
	}
	if mock.gotArgs[2] != "--samples" || mock.gotArgs[3] != "10000" {
		t.Errorf("expected --samples 10000, got %v", mock.gotArgs[2:4])
	}
	if mock.gotArgs[4] != "-o" {
		t.Errorf("expected -o flag, got %v", mock.gotArgs[4])
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.CaptureResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.File == "" {
		t.Error("expected non-empty file name")
	}
}

func TestHandleCaptureDataWithAllOptions(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":       "demo",
		"config":       "samplerate=1M",
		"channels":     "D0,D1,D2",
		"samples":      float64(5000),
		"time":         float64(1000),
		"frames":       float64(10),
		"triggers":     "D0=r",
		"wait_trigger": true,
		"output_file":  "test_capture.sr",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{
		"-d", "demo",
		"-c", "samplerate=1M",
		"-C", "D0,D1,D2",
		"--samples", "5000",
		"--time", "1000",
		"--frames", "10",
		"-t", "D0=r",
		"-w",
		"-o", "test_capture.sr",
	}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.CaptureResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.File != "test_capture.sr" {
		t.Errorf("file = %q, want %q", parsed.File, "test_capture.sr")
	}
}

func TestHandleCaptureDataMissingDriver(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"samples": float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTextResult(t, result, true)
}

func TestHandleCaptureDataMissingSamplesAndTime(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTextResult(t, result, true)
}

func TestHandleCaptureDataInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"invalid driver", map[string]any{"driver": "--evil", "samples": float64(1000)}},
		{"invalid config", map[string]any{"driver": "demo", "samples": float64(1000), "config": ";rm -rf /"}},
		{"invalid channels", map[string]any{"driver": "demo", "samples": float64(1000), "channels": "D0 D1;evil"}},
		{"invalid triggers", map[string]any{"driver": "demo", "samples": float64(1000), "triggers": "$(whoami)"}},
		{"invalid output_file path traversal", map[string]any{"driver": "demo", "samples": float64(1000), "output_file": "../../../etc/passwd"}},
		{"invalid output_file spaces", map[string]any{"driver": "demo", "samples": float64(1000), "output_file": "file name.sr"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleDecodeProtocol(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "uart-1: TX: Start bit\nuart-1: TX: 0x48\nuart-1: TX: Stop bit\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", map[string]any{
		"input_file":        "capture.sr",
		"protocol_decoders": "uart:baudrate=9600",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "capture.sr", "-P", "uart:baudrate=9600"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.DecodeResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Output == "" {
		t.Error("expected non-empty output")
	}
	if parsed.Format != "text" {
		t.Errorf("format = %q, want %q", parsed.Format, "text")
	}
}

func TestHandleDecodeProtocolWithAllOptions(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "decoded output",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", map[string]any{
		"input_file":          "capture.sr",
		"protocol_decoders":   "uart:baudrate=9600",
		"input_format":        "vcd",
		"annotations":         "uart=rx-data",
		"show_sample_numbers": true,
		"meta_output":         "uart=baud",
		"json_trace":          true,
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{
		"-i", "capture.sr",
		"-I", "vcd",
		"-P", "uart:baudrate=9600",
		"-A", "uart=rx-data",
		"--protocol-decoder-samplenum",
		"-M", "uart=baud",
		"--protocol-decoder-jsontrace",
	}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.DecodeResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Format != "json_trace" {
		t.Errorf("format = %q, want %q", parsed.Format, "json_trace")
	}
}

func TestHandleDecodeProtocolMissingParams(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"missing input_file", map[string]any{"protocol_decoders": "uart"}},
		{"missing protocol_decoders", map[string]any{"input_file": "capture.sr"}},
		{"missing both", map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleDecodeProtocolInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"invalid input_file", map[string]any{"input_file": "../evil.sr", "protocol_decoders": "uart"}},
		{"invalid protocol_decoders", map[string]any{"input_file": "capture.sr", "protocol_decoders": ";rm -rf /"}},
		{"invalid input_format", map[string]any{"input_file": "capture.sr", "protocol_decoders": "uart", "input_format": "--evil"}},
		{"invalid annotations", map[string]any{"input_file": "capture.sr", "protocol_decoders": "uart", "annotations": "$(cmd)"}},
		{"invalid meta_output", map[string]any{"input_file": "capture.sr", "protocol_decoders": "uart", "meta_output": "a;b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleCaptureDataTimeOnly(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
		"time":   float64(500),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify --time is present and --samples is not
	if len(mock.gotArgs) < 5 {
		t.Fatalf("expected at least 5 args, got %v", mock.gotArgs)
	}
	if mock.gotArgs[0] != "-d" || mock.gotArgs[1] != "demo" {
		t.Errorf("expected -d demo, got %v", mock.gotArgs[:2])
	}
	if mock.gotArgs[2] != "--time" || mock.gotArgs[3] != "500" {
		t.Errorf("expected --time 500, got %v", mock.gotArgs[2:4])
	}
	for i, arg := range mock.gotArgs {
		if arg == "--samples" {
			t.Errorf("unexpected --samples at position %d", i)
		}
	}

	assertTextResult(t, result, false)
}

func TestHandleCaptureDataNegativeSamples(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "demo",
		"samples": float64(-5),
		"time":    float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := assertTextResult(t, result, true)
	if !contains([]string{text}, "samples must be a positive number") {
		t.Errorf("expected 'samples must be a positive number' error, got %q", text)
	}
}

func TestHandleCaptureDataNegativeTime(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "demo",
		"samples": float64(1000),
		"time":    float64(-5),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := assertTextResult(t, result, true)
	if !contains([]string{text}, "time must be a positive number") {
		t.Errorf("expected 'time must be a positive number' error, got %q", text)
	}
}

func TestHandleCaptureDataOverflowSamples(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "demo",
		"samples": float64(1e16),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTextResult(t, result, true)
}

func TestHandleCaptureDataExecutionError(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		err: errors.New("binary not found"),
	}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "demo",
		"samples": float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text == "" {
		t.Error("expected non-empty error message")
	}
}

func TestHandleCaptureDataNonZeroExit(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stderr:   "Error: device not found",
			ExitCode: 1,
		},
	}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "fx2lafw",
		"samples": float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text != "Error: device not found" {
		t.Errorf("error text = %q, want %q", text, "Error: device not found")
	}
}

func TestHandleCaptureDataNonZeroExitEmptyStderr(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "some output",
			Stderr:   "",
			ExitCode: 1,
		},
	}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "demo",
		"samples": float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text == "" {
		t.Error("expected non-empty error message when stderr is empty")
	}
	if !strings.Contains(text, "exited with code 1") {
		t.Errorf("expected exit code in error, got %q", text)
	}
}

func TestHandleCaptureDataFramesOnly(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
		"frames": float64(10),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify --frames is present and --samples/--time are not
	wantArgs := []string{
		"-d", "demo",
		"--frames", "10",
	}
	for i, want := range wantArgs {
		if i >= len(mock.gotArgs) {
			t.Fatalf("expected at least %d args, got %v", len(wantArgs), mock.gotArgs)
		}
		if mock.gotArgs[i] != want {
			t.Errorf("arg[%d] = %q, want %q", i, mock.gotArgs[i], want)
		}
	}
	for i, arg := range mock.gotArgs {
		if arg == "--samples" {
			t.Errorf("unexpected --samples at position %d", i)
		}
		if arg == "--time" {
			t.Errorf("unexpected --time at position %d", i)
		}
	}

	assertTextResult(t, result, false)
}

func TestHandleCaptureDataFramesAndSamples(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "fx2lafw",
		"frames":  float64(5),
		"samples": float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{
		"-d", "fx2lafw",
		"--samples", "1000",
		"--frames", "5",
	}
	// Check that the expected args are a prefix (before -o)
	for i, want := range wantArgs {
		if i >= len(mock.gotArgs) {
			t.Fatalf("expected at least %d args, got %v", len(wantArgs), mock.gotArgs)
		}
		if mock.gotArgs[i] != want {
			t.Errorf("arg[%d] = %q, want %q", i, mock.gotArgs[i], want)
		}
	}

	assertTextResult(t, result, false)
}

func TestHandleCaptureDataFramesAndTime(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
		"frames": float64(3),
		"time":   float64(500),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{
		"-d", "demo",
		"--time", "500",
		"--frames", "3",
	}
	for i, want := range wantArgs {
		if i >= len(mock.gotArgs) {
			t.Fatalf("expected at least %d args, got %v", len(wantArgs), mock.gotArgs)
		}
		if mock.gotArgs[i] != want {
			t.Errorf("arg[%d] = %q, want %q", i, mock.gotArgs[i], want)
		}
	}

	assertTextResult(t, result, false)
}

func TestHandleCaptureDataNegativeFrames(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
		"frames": float64(-1),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "frames must be a positive number") {
		t.Errorf("expected 'frames must be a positive number' error, got %q", text)
	}
}

func TestHandleCaptureDataOverflowFrames(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
		"frames": float64(1e16),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTextResult(t, result, true)
}

func TestHandleDecodeProtocolExecutionError(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		err: errors.New("binary not found"),
	}, nil, nil, nil)

	result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", map[string]any{
		"input_file":        "capture.sr",
		"protocol_decoders": "uart",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text == "" {
		t.Error("expected non-empty error message")
	}
}

func TestHandleDecodeProtocolNonZeroExit(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stderr:   "Error: input file not found",
			ExitCode: 1,
		},
	}, nil, nil, nil)

	result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", map[string]any{
		"input_file":        "missing.sr",
		"protocol_decoders": "uart",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text != "Error: input file not found" {
		t.Errorf("error text = %q, want %q", text, "Error: input file not found")
	}
}

func TestHandleDecodeProtocolNonZeroExitEmptyStderr(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			Stderr:   "",
			ExitCode: 2,
		},
	}, nil, nil, nil)

	result, err := h.HandleDecodeProtocol(context.Background(), makeRequest("decode_protocol", map[string]any{
		"input_file":        "capture.sr",
		"protocol_decoders": "uart",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "exited with code 2") {
		t.Errorf("expected exit code in error, got %q", text)
	}
}

func TestHandlerNonZeroExit(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			Stderr:   "Error: unknown protocol decoder 'foo'.\n",
			ExitCode: 1,
		},
	}, nil, nil, nil)

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", map[string]any{"decoder": "foo"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

// --- validConnRe tests ---

func TestIsValidConn(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// Valid serial devices
		{"ttyUSB0", "/dev/ttyUSB0", true},
		{"ttyUSB1", "/dev/ttyUSB1", true},
		{"ttyACM0", "/dev/ttyACM0", true},
		{"ttyS0", "/dev/ttyS0", true},
		{"serial with params", "/dev/ttyUSB0:serialcomm=115200/8n1", true},
		{"serial 9600 baud", "/dev/ttyS0:serialcomm=9600/8n1", true},
		{"serial 7e2", "/dev/ttyUSB0:serialcomm=19200/7e2", true},

		// Valid network connections
		{"tcp-raw IP", "tcp-raw/192.168.1.100/5555", true},
		{"tcp IP", "tcp/192.168.1.100/5555", true},
		{"vxi IP", "vxi/192.168.1.100", true},
		{"usbtmc", "usbtmc/1a86.7523", true},

		// Invalid: path traversal
		{"path traversal dot-dot", "/dev/../etc/passwd", false},
		{"tcp path traversal", "tcp/../../../etc/passwd", false},
		{"vxi path traversal", "vxi/192.168.1.100/../../etc/passwd", false},

		// Invalid: shell metacharacters
		{"semicolon", "/dev/ttyUSB0;rm -rf /", false},
		{"backtick", "/dev/ttyUSB0`whoami`", false},
		{"dollar", "/dev/ttyUSB0$(cmd)", false},
		{"pipe", "/dev/ttyUSB0|cat", false},
		{"ampersand", "/dev/ttyUSB0&bg", false},
		{"space", "/dev/tty USB0", false},

		// Invalid: arbitrary paths
		{"etc passwd", "/etc/passwd", false},
		{"home dir", "/home/user/file", false},
		{"tmp file", "/tmp/evil", false},

		// Invalid: empty/malformed
		{"empty string", "", false},
		{"just slash", "/", false},
		{"bare device name", "ttyUSB0", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidConn(tt.input)
			if got != tt.want {
				t.Errorf("isValidConn(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// --- scan_devices with driver/conn tests ---

func TestHandleScanDevicesWithDriverAndConn(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "The following devices were found:\nscpi-dmm - OWON XDM1241\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", map[string]any{
		"driver": "scpi-dmm",
		"conn":   "/dev/ttyUSB0:serialcomm=115200/8n1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-d", "scpi-dmm:conn=/dev/ttyUSB0:serialcomm=115200/8n1", "--scan"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	text := assertTextResult(t, result, false)
	var scanResult sigrok.ScanResult
	if err := json.Unmarshal([]byte(text), &scanResult); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if len(scanResult.Devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(scanResult.Devices))
	}
}

func TestHandleScanDevicesWithDriverOnly(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "The following devices were found:\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	_, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", map[string]any{
		"driver": "scpi-dmm",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-d", "scpi-dmm", "--scan"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}
}

func TestHandleScanDevicesConnWithoutDriver(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", map[string]any{
		"conn": "/dev/ttyUSB0",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "'conn' requires 'driver'") {
		t.Errorf("expected conn-requires-driver error, got %q", text)
	}
}

func TestHandleScanDevicesInvalidConn(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	tests := []struct {
		name string
		conn string
	}{
		{"path traversal", "/dev/../etc/passwd"},
		{"arbitrary path", "/etc/shadow"},
		{"shell injection", "/dev/ttyUSB0;rm -rf /"},
		{"command substitution", "/dev/ttyUSB0$(whoami)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", map[string]any{
				"driver": "scpi-dmm",
				"conn":   tt.conn,
			}))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleScanDevicesInvalidDriver(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", map[string]any{
		"driver": "--evil-flag",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleScanDevicesBackwardCompatibility(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "The following devices were found:\ndemo - Demo device with 13 channels\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	_, err := h.HandleScanDevices(context.Background(), makeRequest("scan_devices", nil))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !reflect.DeepEqual(mock.gotArgs, []string{"--scan"}) {
		t.Errorf("args = %v, want [--scan]", mock.gotArgs)
	}
}

// --- capture_data with conn tests ---

func TestHandleCaptureDataWithConn(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{Stdout: "", ExitCode: 0},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "scpi-dmm",
		"conn":    "/dev/ttyUSB0:serialcomm=115200/8n1",
		"samples": float64(10),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if mock.gotArgs[0] != "-d" {
		t.Fatalf("expected -d, got %s", mock.gotArgs[0])
	}
	if mock.gotArgs[1] != "scpi-dmm:conn=/dev/ttyUSB0:serialcomm=115200/8n1" {
		t.Errorf("driver arg = %q, want %q", mock.gotArgs[1], "scpi-dmm:conn=/dev/ttyUSB0:serialcomm=115200/8n1")
	}

	assertTextResult(t, result, false)
}

func TestHandleCaptureDataInvalidConn(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":  "scpi-dmm",
		"conn":    "/etc/passwd",
		"samples": float64(10),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}
// --- serial_query tests ---

// mockQuerier implements serial.Querier for testing.
type mockQuerier struct {
	result  *serial.QueryResult
	err     error
	gotOpts serial.QueryOptions
}

func (m *mockQuerier) Query(_ context.Context, opts serial.QueryOptions) (*serial.QueryResult, error) {
	m.gotOpts = opts
	return m.result, m.err
}

func TestHandleSerialQueryHappyPath(t *testing.T) {
	mq := &mockQuerier{
		result: &serial.QueryResult{Response: "OWON,XDM1241,1234567,V1.0"},
	}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", map[string]any{
		"port":       "/dev/ttyUSB0",
		"command":    "*IDN?",
		"baudrate":   float64(115200),
		"databits":   float64(8),
		"parity":     "none",
		"stopbits":   "1",
		"timeout_ms": float64(2000),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	if !strings.Contains(text, "OWON,XDM1241") {
		t.Errorf("expected response to contain 'OWON,XDM1241', got %q", text)
	}

	if mq.gotOpts.Port != "/dev/ttyUSB0" {
		t.Errorf("port = %q, want %q", mq.gotOpts.Port, "/dev/ttyUSB0")
	}
	if mq.gotOpts.Command != "*IDN?" {
		t.Errorf("command = %q, want %q", mq.gotOpts.Command, "*IDN?")
	}
	if mq.gotOpts.BaudRate != 115200 {
		t.Errorf("baudrate = %d, want %d", mq.gotOpts.BaudRate, 115200)
	}
	if mq.gotOpts.DataBits != 8 {
		t.Errorf("databits = %d, want %d", mq.gotOpts.DataBits, 8)
	}
	if mq.gotOpts.TimeoutMs != 2000 {
		t.Errorf("timeout_ms = %d, want %d", mq.gotOpts.TimeoutMs, 2000)
	}
}

func TestHandleSerialQueryDefaults(t *testing.T) {
	mq := &mockQuerier{
		result: &serial.QueryResult{Response: "+1.234E+00"},
	}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", map[string]any{
		"port":    "/dev/ttyUSB0",
		"command": "MEAS:VOLT:DC?",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, false)

	// Check defaults were applied
	if mq.gotOpts.BaudRate != 9600 {
		t.Errorf("default baudrate = %d, want %d", mq.gotOpts.BaudRate, 9600)
	}
	if mq.gotOpts.DataBits != 8 {
		t.Errorf("default databits = %d, want %d", mq.gotOpts.DataBits, 8)
	}
	if mq.gotOpts.Parity != "none" {
		t.Errorf("default parity = %q, want %q", mq.gotOpts.Parity, "none")
	}
	if mq.gotOpts.StopBits != "1" {
		t.Errorf("default stopbits = %q, want %q", mq.gotOpts.StopBits, "1")
	}
	if mq.gotOpts.TimeoutMs != 1000 {
		t.Errorf("default timeout_ms = %d, want %d", mq.gotOpts.TimeoutMs, 1000)
	}
}

func TestHandleSerialQueryMissingParams(t *testing.T) {
	mq := &mockQuerier{result: &serial.QueryResult{Response: "ok"}}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"missing port", map[string]any{"command": "*IDN?"}},
		{"missing command", map[string]any{"port": "/dev/ttyUSB0"}},
		{"missing both", map[string]any{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleSerialQueryInvalidPort(t *testing.T) {
	mq := &mockQuerier{result: &serial.QueryResult{Response: "ok"}}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	tests := []struct {
		name string
		port string
	}{
		{"path traversal", "/dev/../etc/passwd"},
		{"arbitrary path", "/etc/passwd"},
		{"shell injection", "/dev/ttyUSB0;rm -rf /"},
		{"command substitution", "/dev/ttyUSB0$(whoami)"},
		{"spaces", "/dev/tty USB0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", map[string]any{
				"port":    tt.port,
				"command": "*IDN?",
			}))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleSerialQueryInvalidCommand(t *testing.T) {
	mq := &mockQuerier{result: &serial.QueryResult{Response: "ok"}}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	tests := []struct {
		name    string
		command string
	}{
		{"shell injection", ";rm -rf /"},
		{"command substitution", "$(whoami)"},
		{"backtick", "`whoami`"},
		{"pipe", "IDN?|cat"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", map[string]any{
				"port":    "/dev/ttyUSB0",
				"command": tt.command,
			}))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleSerialQueryInvalidNumericParams(t *testing.T) {
	mq := &mockQuerier{result: &serial.QueryResult{Response: "ok"}}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"negative baudrate", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "baudrate": float64(-1)}},
		{"zero baudrate", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "baudrate": float64(0)}},
		{"databits too low", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "databits": float64(4)}},
		{"databits too high", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "databits": float64(9)}},
		{"negative timeout", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "timeout_ms": float64(-1)}},
		{"zero timeout", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "timeout_ms": float64(0)}},
		{"timeout too large", map[string]any{"port": "/dev/ttyUSB0", "command": "*IDN?", "timeout_ms": float64(60000)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleSerialQueryDeviceError(t *testing.T) {
	mq := &mockQuerier{
		err: errors.New("open port /dev/ttyUSB0: permission denied"),
	}
	h := NewHandlers(&mockExecutor{}, nil, mq, nil)

	result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", map[string]any{
		"port":    "/dev/ttyUSB0",
		"command": "*IDN?",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "serial query failed") {
		t.Errorf("expected 'serial query failed' error, got %q", text)
	}
}

func TestHandleSerialQueryNilQuerier(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleSerialQuery(context.Background(), makeRequest("serial_query", map[string]any{
		"port":    "/dev/ttyUSB0",
		"command": "*IDN?",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "not available") {
		t.Errorf("expected 'not available' error, got %q", text)
	}
}

// --- get_device_profile tests ---

func testDeviceRegistry() *devices.Registry {
	return devices.NewRegistry([]*devices.Profile{
		{
			ID:           "owon-xdm1241",
			Manufacturer: "OWON",
			Model:        "XDM1241",
			Description:  "4 1/2 digit bench-type digital multimeter",
			IDNPattern:   "OWON,XDM1241,",
			Connection: devices.Connection{
				BaudRate:  115200,
				DataBits:  8,
				Parity:    "none",
				StopBits:  "1",
				TimeoutMs: 3000,
			},
			Commands: []devices.Command{
				{Name: "*IDN?", Description: "Device identification", ExampleResponse: "OWON,XDM1241,24412417,V4.3.0,3"},
			},
			Notes: []string{"Uses non-standard flat SCPI command set"},
		},
	})
}

func TestHandleGetDeviceProfileHappyPath(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, testDeviceRegistry())

	result, err := h.HandleGetDeviceProfile(context.Background(), makeRequest("get_device_profile", map[string]any{
		"query": "XDM1241",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed struct {
		Query   string `json:"query"`
		Count   int    `json:"count"`
		Matches []struct {
			ID    string `json:"id"`
			Model string `json:"model"`
		} `json:"matches"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Count != 1 {
		t.Fatalf("count = %d, want 1", parsed.Count)
	}
	if parsed.Matches[0].ID != "owon-xdm1241" {
		t.Errorf("match ID = %q, want %q", parsed.Matches[0].ID, "owon-xdm1241")
	}
	if parsed.Query != "XDM1241" {
		t.Errorf("query = %q, want %q", parsed.Query, "XDM1241")
	}
}

func TestHandleGetDeviceProfileIDNMatch(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, testDeviceRegistry())

	result, err := h.HandleGetDeviceProfile(context.Background(), makeRequest("get_device_profile", map[string]any{
		"query": "OWON,XDM1241,24412417,V4.3.0,3",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed struct {
		Count   int `json:"count"`
		Matches []struct {
			ID string `json:"id"`
		} `json:"matches"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Count != 1 {
		t.Fatalf("count = %d, want 1", parsed.Count)
	}
	if parsed.Matches[0].ID != "owon-xdm1241" {
		t.Errorf("match ID = %q, want %q", parsed.Matches[0].ID, "owon-xdm1241")
	}
}

func TestHandleGetDeviceProfileNoMatch(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, testDeviceRegistry())

	result, err := h.HandleGetDeviceProfile(context.Background(), makeRequest("get_device_profile", map[string]any{
		"query": "Keysight",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Count != 0 {
		t.Errorf("count = %d, want 0", parsed.Count)
	}
}

func TestHandleGetDeviceProfileMissingQuery(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, testDeviceRegistry())

	result, err := h.HandleGetDeviceProfile(context.Background(), makeRequest("get_device_profile", map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}

func TestHandleGetDeviceProfileInvalidQuery(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, testDeviceRegistry())

	tests := []struct {
		name  string
		query string
	}{
		{"shell injection", ";rm -rf /"},
		{"command substitution", "$(whoami)"},
		{"backtick", "`whoami`"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleGetDeviceProfile(context.Background(), makeRequest("get_device_profile", map[string]any{
				"query": tt.query,
			}))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleGetDeviceProfileNilRegistry(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleGetDeviceProfile(context.Background(), makeRequest("get_device_profile", map[string]any{
		"query": "XDM1241",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "not available") {
		t.Errorf("expected 'not available' error, got %q", text)
	}
}

// --- render_waveform tests ---

func TestHandleRenderWaveformASCIIDefault(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "D0:11110000\nD1:00001111\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file": "capture.sr",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "capture.sr", "-O", "ascii"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.RenderResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Output == "" {
		t.Error("expected non-empty output")
	}
	if parsed.Format != "ascii" {
		t.Errorf("format = %q, want %q", parsed.Format, "ascii")
	}
}

func TestHandleRenderWaveformASCIIExplicit(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "D0:11110000\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file":    "capture.sr",
		"output_format": "ascii",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "capture.sr", "-O", "ascii"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	assertTextResult(t, result, false)
}

func TestHandleRenderWaveformWaveDrom(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   `{"signal":[{"name":"D0","wave":"10101"}]}`,
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file":    "capture.sr",
		"output_format": "wavedrom",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "capture.sr", "-O", "wavedrom"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.RenderResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Format != "wavedrom" {
		t.Errorf("format = %q, want %q", parsed.Format, "wavedrom")
	}
}

func TestHandleRenderWaveformWithInputFormat(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "waveform output",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file":   "data.csv",
		"input_format": "csv",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "data.csv", "-I", "csv", "-O", "ascii"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	assertTextResult(t, result, false)
}

func TestHandleRenderWaveformWithChannels(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "D0:11110000\nD1:00001111\n",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file": "capture.sr",
		"channels":   "D0,D1",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "capture.sr", "-O", "ascii", "-C", "D0,D1"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	assertTextResult(t, result, false)
}

func TestHandleRenderWaveformAllParams(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   `{"signal":[]}`,
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file":    "data.csv",
		"output_format": "wavedrom",
		"input_format":  "csv",
		"channels":      "D0,D1,D2",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantArgs := []string{"-i", "data.csv", "-I", "csv", "-O", "wavedrom", "-C", "D0,D1,D2"}
	if !reflect.DeepEqual(mock.gotArgs, wantArgs) {
		t.Errorf("args = %v, want %v", mock.gotArgs, wantArgs)
	}

	assertTextResult(t, result, false)
}

func TestHandleRenderWaveformMissingInputFile(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "missing required parameter: input_file") {
		t.Errorf("expected missing param error, got %q", text)
	}
}

func TestHandleRenderWaveformInvalidParams(t *testing.T) {
	h := NewHandlers(&mockExecutor{}, nil, nil, nil)

	tests := []struct {
		name string
		args map[string]any
	}{
		{"path traversal input_file", map[string]any{"input_file": "../etc/passwd"}},
		{"input_file with slash", map[string]any{"input_file": "dir/file.sr"}},
		{"invalid output_format", map[string]any{"input_file": "capture.sr", "output_format": "binary"}},
		{"unsupported output_format csv", map[string]any{"input_file": "capture.sr", "output_format": "csv"}},
		{"unsupported output_format hex", map[string]any{"input_file": "capture.sr", "output_format": "hex"}},
		{"invalid input_format", map[string]any{"input_file": "capture.sr", "input_format": "--evil"}},
		{"invalid channels", map[string]any{"input_file": "capture.sr", "channels": "$(cmd)"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", tt.args))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assertTextResult(t, result, true)
		})
	}
}

func TestHandleRenderWaveformExecutionError(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		err: errors.New("binary not found"),
	}, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file": "capture.sr",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "sigrok-cli execution failed") {
		t.Errorf("expected execution error, got %q", text)
	}
}

func TestHandleRenderWaveformNonZeroExit(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stderr:   "Error: input file not found",
			ExitCode: 1,
		},
	}, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file": "missing.sr",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if text != "Error: input file not found" {
		t.Errorf("error text = %q, want %q", text, "Error: input file not found")
	}
}

func TestHandleRenderWaveformNonZeroExitEmptyStderr(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			Stderr:   "",
			ExitCode: 1,
		},
	}, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file": "capture.sr",
	}))
	if err != nil {
		t.Fatalf("unexpected Go error: %v", err)
	}

	text := assertTextResult(t, result, true)
	if !strings.Contains(text, "exited with code 1") {
		t.Errorf("expected exit code in error, got %q", text)
	}
}

func TestHandleRenderWaveformEmptyStdout(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock, nil, nil, nil)

	result, err := h.HandleRenderWaveform(context.Background(), makeRequest("render_waveform", map[string]any{
		"input_file": "empty.sr",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	text := assertTextResult(t, result, false)
	var parsed sigrok.RenderResult
	if err := json.Unmarshal([]byte(text), &parsed); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	if parsed.Output != "" {
		t.Errorf("expected empty output, got %q", parsed.Output)
	}
}
