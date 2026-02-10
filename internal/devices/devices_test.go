package devices

import (
	"testing"
)

func TestLoadEmbedded(t *testing.T) {
	reg, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("LoadEmbedded() error: %v", err)
	}

	profiles := reg.List()
	if len(profiles) == 0 {
		t.Fatal("expected at least one embedded profile")
	}

	// Verify the OWON profile loaded correctly.
	p, ok := reg.Get("owon-xdm1241")
	if !ok {
		t.Fatal("expected owon-xdm1241 profile")
	}
	if p.Manufacturer != "OWON" {
		t.Errorf("manufacturer = %q, want %q", p.Manufacturer, "OWON")
	}
	if p.Model != "XDM1241" {
		t.Errorf("model = %q, want %q", p.Model, "XDM1241")
	}
	if p.Connection.BaudRate != 115200 {
		t.Errorf("baudrate = %d, want %d", p.Connection.BaudRate, 115200)
	}
	if p.Connection.TimeoutMs != 3000 {
		t.Errorf("timeout_ms = %d, want %d", p.Connection.TimeoutMs, 3000)
	}
	if len(p.Commands) == 0 {
		t.Error("expected at least one command")
	}
	if len(p.Notes) == 0 {
		t.Error("expected at least one note")
	}
	if p.IDNPattern != "OWON,XDM1241," {
		t.Errorf("idn_pattern = %q, want %q", p.IDNPattern, "OWON,XDM1241,")
	}
}

func TestNewRegistry(t *testing.T) {
	profiles := []*Profile{
		{ID: "device-a", Manufacturer: "MakerA", Model: "ModelA"},
		{ID: "device-b", Manufacturer: "MakerB", Model: "ModelB", IDNPattern: "MakerB,ModelB,"},
	}
	reg := NewRegistry(profiles)

	if got := len(reg.List()); got != 2 {
		t.Errorf("List() length = %d, want 2", got)
	}

	p, ok := reg.Get("device-a")
	if !ok {
		t.Fatal("expected device-a")
	}
	if p.Manufacturer != "MakerA" {
		t.Errorf("manufacturer = %q, want %q", p.Manufacturer, "MakerA")
	}
}

func TestRegistryGet(t *testing.T) {
	reg := NewRegistry([]*Profile{
		{ID: "test-device", Manufacturer: "Test", Model: "Dev1"},
	})

	t.Run("exact match", func(t *testing.T) {
		p, ok := reg.Get("test-device")
		if !ok {
			t.Fatal("expected to find test-device")
		}
		if p.ID != "test-device" {
			t.Errorf("ID = %q, want %q", p.ID, "test-device")
		}
	})

	t.Run("case insensitive", func(t *testing.T) {
		p, ok := reg.Get("TEST-DEVICE")
		if !ok {
			t.Fatal("expected case-insensitive match")
		}
		if p.ID != "test-device" {
			t.Errorf("ID = %q, want %q", p.ID, "test-device")
		}
	})

	t.Run("missing", func(t *testing.T) {
		_, ok := reg.Get("nonexistent")
		if ok {
			t.Error("expected no match for nonexistent ID")
		}
	})
}

func TestRegistryList(t *testing.T) {
	reg := NewRegistry([]*Profile{
		{ID: "device-b"},
		{ID: "device-a"},
	})

	profiles := reg.List()
	if len(profiles) != 2 {
		t.Fatalf("List() length = %d, want 2", len(profiles))
	}
	if profiles[0].ID != "device-a" {
		t.Errorf("first profile ID = %q, want %q", profiles[0].ID, "device-a")
	}
	if profiles[1].ID != "device-b" {
		t.Errorf("second profile ID = %q, want %q", profiles[1].ID, "device-b")
	}
}

func TestRegistryLookup(t *testing.T) {
	reg := NewRegistry([]*Profile{
		{ID: "owon-xdm1241", Manufacturer: "OWON", Model: "XDM1241", IDNPattern: "OWON,XDM1241,"},
		{ID: "rigol-dm3068", Manufacturer: "Rigol", Model: "DM3068", IDNPattern: "Rigol,DM3068,"},
	})

	tests := []struct {
		name      string
		query     string
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "exact ID match",
			query:     "owon-xdm1241",
			wantCount: 1,
			wantIDs:   []string{"owon-xdm1241"},
		},
		{
			name:      "case insensitive ID",
			query:     "OWON-XDM1241",
			wantCount: 1,
			wantIDs:   []string{"owon-xdm1241"},
		},
		{
			name:      "model match",
			query:     "XDM1241",
			wantCount: 1,
			wantIDs:   []string{"owon-xdm1241"},
		},
		{
			name:      "manufacturer match",
			query:     "Rigol",
			wantCount: 1,
			wantIDs:   []string{"rigol-dm3068"},
		},
		{
			name:      "IDN response match",
			query:     "OWON,XDM1241,24412417,V4.3.0,3",
			wantCount: 1,
			wantIDs:   []string{"owon-xdm1241"},
		},
		{
			name:      "no match",
			query:     "Keysight",
			wantCount: 0,
		},
		{
			name:      "empty query",
			query:     "",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			matches := reg.Lookup(tt.query)
			if len(matches) != tt.wantCount {
				t.Fatalf("Lookup(%q) returned %d matches, want %d", tt.query, len(matches), tt.wantCount)
			}
			for i, wantID := range tt.wantIDs {
				if matches[i].ID != wantID {
					t.Errorf("match[%d].ID = %q, want %q", i, matches[i].ID, wantID)
				}
			}
		})
	}
}
