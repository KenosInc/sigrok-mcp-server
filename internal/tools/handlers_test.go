package tools

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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

func TestHandleCaptureData(t *testing.T) {
	mock := &mockExecutor{
		result: &sigrok.CommandResult{
			Stdout:   "",
			ExitCode: 0,
		},
	}
	h := NewHandlers(mock)

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
	h := NewHandlers(mock)

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver":       "demo",
		"config":       "samplerate=1M",
		"channels":     "D0,D1,D2",
		"samples":      float64(5000),
		"time":         float64(1000),
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
	h := NewHandlers(&mockExecutor{})

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"samples": float64(1000),
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTextResult(t, result, true)
}

func TestHandleCaptureDataMissingSamplesAndTime(t *testing.T) {
	h := NewHandlers(&mockExecutor{})

	result, err := h.HandleCaptureData(context.Background(), makeRequest("capture_data", map[string]any{
		"driver": "demo",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertTextResult(t, result, true)
}

func TestHandleCaptureDataInvalidParam(t *testing.T) {
	h := NewHandlers(&mockExecutor{})

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
	h := NewHandlers(mock)

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
	h := NewHandlers(mock)

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
	h := NewHandlers(&mockExecutor{})

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
	h := NewHandlers(&mockExecutor{})

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
	h := NewHandlers(mock)

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
	h := NewHandlers(&mockExecutor{})

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
	h := NewHandlers(&mockExecutor{})

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
	h := NewHandlers(&mockExecutor{})

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
	})

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
	})

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
	})

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

func TestHandleDecodeProtocolExecutionError(t *testing.T) {
	h := NewHandlers(&mockExecutor{
		err: errors.New("binary not found"),
	})

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
	})

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
	})

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
	})

	result, err := h.HandleShowDecoderDetails(context.Background(), makeRequest("show_decoder_details", map[string]any{"decoder": "foo"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertTextResult(t, result, true)
}
