package devices

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed profiles/*.json
var profilesFS embed.FS

// Profile represents a device profile with connection settings and command reference.
type Profile struct {
	ID           string     `json:"id"`
	Manufacturer string     `json:"manufacturer"`
	Model        string     `json:"model"`
	Description  string     `json:"description"`
	IDNPattern   string     `json:"idn_pattern,omitempty"`
	Connection   Connection `json:"connection"`
	Commands     []Command  `json:"commands,omitempty"`
	Notes        []string   `json:"notes,omitempty"`
}

// Connection holds serial port settings for a device.
type Connection struct {
	BaudRate  int    `json:"baudrate"`
	DataBits  int    `json:"databits"`
	Parity    string `json:"parity"`
	StopBits  string `json:"stopbits"`
	TimeoutMs int    `json:"timeout_ms"`
}

// Command describes a device command with an optional example response.
type Command struct {
	Name            string `json:"name"`
	Description     string `json:"description"`
	ExampleResponse string `json:"example_response,omitempty"`
}

// Registry holds loaded device profiles and supports lookup.
type Registry struct {
	profiles []*Profile
	byID     map[string]*Profile
}

// LoadEmbedded reads all embedded profile JSON files and returns a Registry.
func LoadEmbedded() (*Registry, error) {
	entries, err := profilesFS.ReadDir("profiles")
	if err != nil {
		return nil, fmt.Errorf("read embedded profiles: %w", err)
	}

	var profiles []*Profile
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := profilesFS.ReadFile("profiles/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read profile %s: %w", e.Name(), err)
		}
		var p Profile
		if err := json.Unmarshal(data, &p); err != nil {
			return nil, fmt.Errorf("parse profile %s: %w", e.Name(), err)
		}
		profiles = append(profiles, &p)
	}

	return NewRegistry(profiles), nil
}

// NewRegistry creates a Registry from a slice of profiles.
func NewRegistry(profiles []*Profile) *Registry {
	byID := make(map[string]*Profile, len(profiles))
	for _, p := range profiles {
		byID[strings.ToLower(p.ID)] = p
	}
	return &Registry{profiles: profiles, byID: byID}
}

// Get returns a profile by exact ID (case-insensitive).
func (r *Registry) Get(id string) (*Profile, bool) {
	p, ok := r.byID[strings.ToLower(id)]
	return p, ok
}

// List returns all profiles sorted by ID.
func (r *Registry) List() []*Profile {
	sorted := make([]*Profile, len(r.profiles))
	copy(sorted, r.profiles)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].ID < sorted[j].ID
	})
	return sorted
}

// Lookup finds profiles matching a query string. It checks against ID, model,
// manufacturer, and IDN pattern (in that order). An exact ID match returns
// immediately with a single result.
func (r *Registry) Lookup(query string) []*Profile {
	if query == "" {
		return nil
	}
	lower := strings.ToLower(query)

	// Exact ID match
	if p, ok := r.byID[lower]; ok {
		return []*Profile{p}
	}

	var matches []*Profile
	seen := make(map[string]bool)

	// Model match (case-insensitive substring)
	for _, p := range r.profiles {
		if strings.Contains(lower, strings.ToLower(p.Model)) {
			if !seen[p.ID] {
				matches = append(matches, p)
				seen[p.ID] = true
			}
		}
	}

	// Manufacturer match (case-insensitive substring)
	for _, p := range r.profiles {
		if strings.Contains(lower, strings.ToLower(p.Manufacturer)) {
			if !seen[p.ID] {
				matches = append(matches, p)
				seen[p.ID] = true
			}
		}
	}

	// IDN pattern match (check if query contains the idn_pattern)
	for _, p := range r.profiles {
		if p.IDNPattern != "" && strings.Contains(lower, strings.ToLower(p.IDNPattern)) {
			if !seen[p.ID] {
				matches = append(matches, p)
				seen[p.ID] = true
			}
		}
	}

	return matches
}
