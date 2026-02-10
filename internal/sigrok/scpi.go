package sigrok

// SCPICategory describes a category of SCPI-capable instruments
// along with their sigrok driver IDs and typical connection methods.
type SCPICategory struct {
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Drivers     []string `json:"drivers"`
	ConnTypes   []string `json:"conn_types"`
}

// SCPICategories returns all known SCPI device categories with their drivers.
// Adding a new device category requires only appending an entry here.
func SCPICategories() []SCPICategory {
	return []SCPICategory{
		{
			Category:    "DMM",
			Description: "Digital Multimeters (SCPI)",
			Drivers:     []string{"scpi-dmm"},
			ConnTypes:   []string{"serial", "usbtmc", "tcp"},
		},
		{
			Category:    "Power Supply",
			Description: "Programmable Power Supplies (SCPI)",
			Drivers:     []string{"scpi-pps"},
			ConnTypes:   []string{"serial", "usbtmc", "tcp"},
		},
		{
			Category:    "Oscilloscope",
			Description: "Oscilloscopes (vendor-specific SCPI drivers)",
			Drivers:     []string{"rigol-ds", "siglent-sds", "lecroy-xstream", "hameg-hmo", "yokogawa-dlm", "gwinstek-gds-800"},
			ConnTypes:   []string{"tcp", "usbtmc", "vxi"},
		},
		{
			Category:    "Signal Generator",
			Description: "Signal/Function Generators",
			Drivers:     []string{"rigol-dg"},
			ConnTypes:   []string{"tcp", "usbtmc"},
		},
	}
}
