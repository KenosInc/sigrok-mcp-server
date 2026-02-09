package sigrok

import (
	"os"
	"path/filepath"
	"testing"
)

func readTestdata(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("failed to read testdata/%s: %v", name, err)
	}
	return string(data)
}

func TestParseListSection(t *testing.T) {
	fullOutput := readTestdata(t, "sigrok_cli_L.txt")

	tests := []struct {
		name        string
		section     string
		wantMinLen  int
		wantFirstID string
		wantFirstDesc string
	}{
		{
			name:          "hardware drivers",
			section:       "Supported hardware drivers:",
			wantMinLen:    10,
			wantFirstID:   "agilent-dmm",
			wantFirstDesc: "Agilent U12xx series DMMs",
		},
		{
			name:          "input formats",
			section:       "Supported input formats:",
			wantMinLen:    5,
			wantFirstID:   "binary",
			wantFirstDesc: "Raw binary logic data",
		},
		{
			name:          "output formats",
			section:       "Supported output formats:",
			wantMinLen:    5,
			wantFirstID:   "analog",
			wantFirstDesc: "ASCII analog data values and units",
		},
		{
			name:          "protocol decoders",
			section:       "Supported protocol decoders:",
			wantMinLen:    10,
			wantFirstID:   "ac97",
			wantFirstDesc: "Audio Codec '97",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items := ParseListSection(fullOutput, tt.section)

			if len(items) < tt.wantMinLen {
				t.Fatalf("got %d items, want at least %d", len(items), tt.wantMinLen)
			}
			if items[0].ID != tt.wantFirstID {
				t.Errorf("first item ID = %q, want %q", items[0].ID, tt.wantFirstID)
			}
			if items[0].Description != tt.wantFirstDesc {
				t.Errorf("first item Description = %q, want %q", items[0].Description, tt.wantFirstDesc)
			}
		})
	}
}

func TestParseListSectionSingleSpaceSeparator(t *testing.T) {
	// Regression: some sigrok-cli lines use only a single space between ID and description.
	input := "Supported hardware drivers:\n  arachnid-labs-re-load-pro Arachnid Labs Re:load Pro\n  demo                 Demo driver\n"
	items := ParseListSection(input, "Supported hardware drivers:")
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != "arachnid-labs-re-load-pro" {
		t.Errorf("first item ID = %q, want %q", items[0].ID, "arachnid-labs-re-load-pro")
	}
	if items[0].Description != "Arachnid Labs Re:load Pro" {
		t.Errorf("first item Description = %q, want %q", items[0].Description, "Arachnid Labs Re:load Pro")
	}
}

func TestParseListSectionUnknown(t *testing.T) {
	items := ParseListSection("some random text", "Nonexistent section:")
	if len(items) != 0 {
		t.Errorf("expected empty result for unknown section, got %d items", len(items))
	}
}

func TestParseDecoderDetails(t *testing.T) {
	output := readTestdata(t, "show_decoder_uart.txt")
	details := ParseDecoderDetails(output)

	if details.ID != "uart" {
		t.Errorf("ID = %q, want %q", details.ID, "uart")
	}
	if details.Name != "UART" {
		t.Errorf("Name = %q, want %q", details.Name, "UART")
	}
	if details.LongName != "Universal Asynchronous Receiver/Transmitter" {
		t.Errorf("LongName = %q, want %q", details.LongName, "Universal Asynchronous Receiver/Transmitter")
	}
	if details.Description != "Asynchronous, serial bus." {
		t.Errorf("Description = %q, want %q", details.Description, "Asynchronous, serial bus.")
	}
	if details.License != "gplv2+" {
		t.Errorf("License = %q, want %q", details.License, "gplv2+")
	}
	if len(details.Inputs) != 1 || details.Inputs[0] != "logic" {
		t.Errorf("Inputs = %v, want [logic]", details.Inputs)
	}
	if len(details.Outputs) != 1 || details.Outputs[0] != "uart" {
		t.Errorf("Outputs = %v, want [uart]", details.Outputs)
	}
	if len(details.Options) == 0 {
		t.Error("expected at least one option")
	}
	if details.Documentation == "" {
		t.Error("expected non-empty documentation")
	}
}

func TestParseDriverDetails(t *testing.T) {
	output := readTestdata(t, "show_driver_demo.txt")
	details := ParseDriverDetails(output)

	if len(details.Functions) == 0 {
		t.Error("expected at least one function")
	}
	foundDemo := false
	for _, f := range details.Functions {
		if f == "Demo device" {
			foundDemo = true
		}
	}
	if !foundDemo {
		t.Errorf("expected 'Demo device' in functions, got %v", details.Functions)
	}
	if len(details.ScanOptions) == 0 {
		t.Error("expected at least one scan option")
	}
	if len(details.Devices) == 0 {
		t.Error("expected at least one device")
	}
}

func TestParseVersion(t *testing.T) {
	output := readTestdata(t, "version.txt")
	info := ParseVersion(output)

	if info.CLIVersion != "0.7.2" {
		t.Errorf("CLIVersion = %q, want %q", info.CLIVersion, "0.7.2")
	}
	if info.LibsigrokVersion == "" {
		t.Error("expected non-empty LibsigrokVersion")
	}
	if info.LibsigrokdecodeVersion == "" {
		t.Error("expected non-empty LibsigrokdecodeVersion")
	}
}

func TestParseScanDevices(t *testing.T) {
	output := readTestdata(t, "scan.txt")
	devices := ParseScanDevices(output)

	if len(devices) == 0 {
		t.Fatal("expected at least one device")
	}

	found := false
	for _, d := range devices {
		if d.Driver == "demo" && d.Description == "Demo device with 13 channels: D0 D1 D2 D3 D4 D5 D6 D7 A0 A1 A2 A3 A4" {
			found = true
		}
	}
	if !found {
		t.Errorf("expected demo device, got %v", devices)
	}
}

func TestParseScanDevicesEmpty(t *testing.T) {
	devices := ParseScanDevices("")
	if len(devices) != 0 {
		t.Errorf("expected empty result for empty input, got %d devices", len(devices))
	}
}
