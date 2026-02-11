package sigrok

import "testing"

func TestSCPICategories(t *testing.T) {
	categories := SCPICategories()

	if len(categories) == 0 {
		t.Fatal("expected at least one SCPI category")
	}

	for _, cat := range categories {
		t.Run(cat.Category, func(t *testing.T) {
			if cat.Category == "" {
				t.Error("category name must not be empty")
			}
			if cat.Description == "" {
				t.Error("description must not be empty")
			}
			if len(cat.Drivers) == 0 {
				t.Error("expected at least one driver")
			}
			if len(cat.ConnTypes) == 0 {
				t.Error("expected at least one connection type")
			}
		})
	}
}

func TestSCPICategoriesContainExpectedDrivers(t *testing.T) {
	categories := SCPICategories()

	// Build a map of category -> drivers for easy lookup.
	catDrivers := make(map[string][]string)
	for _, cat := range categories {
		catDrivers[cat.Category] = cat.Drivers
	}

	tests := []struct {
		category string
		driver   string
	}{
		{"DMM", "scpi-dmm"},
		{"Power Supply", "scpi-pps"},
		{"Oscilloscope", "rigol-ds"},
		{"Oscilloscope", "siglent-sds"},
		{"Signal Generator", "rigol-dg"},
	}
	for _, tt := range tests {
		t.Run(tt.category+"/"+tt.driver, func(t *testing.T) {
			drivers, ok := catDrivers[tt.category]
			if !ok {
				t.Fatalf("category %q not found", tt.category)
			}
			found := false
			for _, d := range drivers {
				if d == tt.driver {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("driver %q not found in category %q (has %v)", tt.driver, tt.category, drivers)
			}
		})
	}
}
