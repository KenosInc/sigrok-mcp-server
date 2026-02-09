package sigrok

import (
	"regexp"
	"strings"
)

// ListItem represents a single entry from a sigrok-cli -L section.
type ListItem struct {
	ID          string `json:"id"`
	Description string `json:"description"`
}

// DecoderDetails holds parsed output from sigrok-cli --show -P <decoder>.
type DecoderDetails struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	LongName      string   `json:"long_name"`
	Description   string   `json:"description"`
	License       string   `json:"license"`
	Inputs        []string `json:"inputs"`
	Outputs       []string `json:"outputs"`
	Options       []string `json:"options"`
	Documentation string   `json:"documentation"`
}

// DriverDetails holds parsed output from sigrok-cli --show -d <driver>.
type DriverDetails struct {
	Functions   []string `json:"functions"`
	ScanOptions []string `json:"scan_options"`
	Devices     []string `json:"devices"`
	RawOutput   string   `json:"raw_output"`
}

// VersionInfo holds parsed output from sigrok-cli --version.
type VersionInfo struct {
	CLIVersion              string `json:"cli_version"`
	LibsigrokVersion        string `json:"libsigrok_version"`
	LibsigrokdecodeVersion  string `json:"libsigrokdecode_version"`
	RawOutput               string `json:"raw_output"`
}

// ScannedDevice holds a single device from sigrok-cli --scan.
type ScannedDevice struct {
	Driver      string `json:"driver"`
	Description string `json:"description"`
}

// ParseListSection extracts items from a named section of sigrok-cli -L output.
// sectionHeader should match the full header line, e.g. "Supported hardware drivers:".
func ParseListSection(output, sectionHeader string) []ListItem {
	lines := strings.Split(output, "\n")
	var items []ListItem
	inSection := false

	for _, line := range lines {
		if strings.TrimSpace(line) == sectionHeader {
			inSection = true
			continue
		}
		if inSection {
			// A new section starts with a non-indented line
			if line != "" && !strings.HasPrefix(line, "  ") {
				break
			}
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			id, desc := splitListLine(trimmed)
			if id != "" {
				items = append(items, ListItem{ID: id, Description: desc})
			}
		}
	}

	return items
}

// splitListLine splits a line like "  agilent-dmm          Agilent U12xx series DMMs"
// into ID and description parts.
var multiSpaceRe = regexp.MustCompile(`\s{2,}`)

func splitListLine(line string) (string, string) {
	// Split on 2+ whitespace characters to separate ID from description.
	loc := multiSpaceRe.FindStringIndex(line)
	if loc != nil {
		return strings.TrimSpace(line[:loc[0]]), strings.TrimSpace(line[loc[1]:])
	}
	return strings.TrimSpace(line), ""
}

// ParseDecoderDetails parses the output of sigrok-cli --show -P <decoder>.
func ParseDecoderDetails(output string) DecoderDetails {
	var d DecoderDetails
	lines := strings.Split(output, "\n")

	inDoc := false
	var docLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		if inDoc {
			docLines = append(docLines, line)
			continue
		}

		if strings.HasPrefix(line, "ID: ") {
			d.ID = strings.TrimPrefix(line, "ID: ")
		} else if strings.HasPrefix(line, "Name: ") {
			d.Name = strings.TrimPrefix(line, "Name: ")
		} else if strings.HasPrefix(line, "Long name: ") {
			d.LongName = strings.TrimPrefix(line, "Long name: ")
		} else if strings.HasPrefix(line, "Description: ") {
			d.Description = strings.TrimPrefix(line, "Description: ")
		} else if strings.HasPrefix(line, "License: ") {
			d.License = strings.TrimPrefix(line, "License: ")
		} else if line == "Possible decoder input IDs:" {
			d.Inputs = collectDashItems(lines, i+1)
		} else if line == "Possible decoder output IDs:" {
			d.Outputs = collectDashItems(lines, i+1)
		} else if strings.HasPrefix(line, "Options:") {
			d.Options = collectDashItems(lines, i+1)
		} else if line == "Documentation:" {
			inDoc = true
		}
	}

	d.Documentation = strings.TrimSpace(strings.Join(docLines, "\n"))
	return d
}

// collectDashItems collects lines starting with "- " from startIdx onward,
// stopping at the first non-matching non-empty line.
func collectDashItems(lines []string, startIdx int) []string {
	var items []string
	for i := startIdx; i < len(lines); i++ {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "- ") {
			break
		}
		items = append(items, strings.TrimPrefix(trimmed, "- "))
	}
	return items
}

// ParseDriverDetails parses the output of sigrok-cli --show -d <driver>.
func ParseDriverDetails(output string) DriverDetails {
	d := DriverDetails{RawOutput: output}
	lines := strings.Split(output, "\n")

	section := ""
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		isIndented := strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t")

		switch {
		case trimmed == "Driver functions:":
			section = "functions"
		case trimmed == "Scan options:":
			section = "scan_options"
		case !isIndented && strings.Contains(trimmed, " - "):
			// Non-indented line with " - " separator is a device line.
			section = "devices"
			d.Devices = append(d.Devices, trimmed)
		case !isIndented:
			// Non-indented, non-device line ends the current section.
			section = ""
		default:
			switch section {
			case "functions":
				d.Functions = append(d.Functions, trimmed)
			case "scan_options":
				d.ScanOptions = append(d.ScanOptions, trimmed)
			}
		}
	}

	return d
}

// ParseVersion parses the output of sigrok-cli --version.
func ParseVersion(output string) VersionInfo {
	info := VersionInfo{RawOutput: output}

	// First line: "sigrok-cli 0.7.2"
	lines := strings.Split(output, "\n")
	if len(lines) > 0 {
		parts := strings.SplitN(lines[0], " ", 2)
		if len(parts) == 2 {
			info.CLIVersion = strings.TrimSpace(parts[1])
		}
	}

	// Look for libsigrok and libsigrokdecode versions
	libsigrokRe := regexp.MustCompile(`libsigrok (\S+)`)
	libsigrokdecodeRe := regexp.MustCompile(`libsigrokdecode (\S+)`)

	for _, line := range lines {
		if m := libsigrokRe.FindStringSubmatch(line); m != nil && info.LibsigrokVersion == "" {
			info.LibsigrokVersion = m[1]
		}
		if m := libsigrokdecodeRe.FindStringSubmatch(line); m != nil && info.LibsigrokdecodeVersion == "" {
			info.LibsigrokdecodeVersion = m[1]
		}
	}

	return info
}

// ParseScanDevices parses the output of sigrok-cli --scan.
func ParseScanDevices(output string) []ScannedDevice {
	var devices []ScannedDevice
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "The following") {
			continue
		}
		// Device lines: "driver_id - Description text"
		parts := strings.SplitN(trimmed, " - ", 2)
		if len(parts) == 2 {
			devices = append(devices, ScannedDevice{
				Driver:      strings.TrimSpace(parts[0]),
				Description: strings.TrimSpace(parts[1]),
			})
		}
	}

	return devices
}
