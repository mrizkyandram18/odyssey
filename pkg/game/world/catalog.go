package world

import (
	"context"
	"encoding/json"
	"fmt"
)

// RealmDefinition holds the configurable metadata for a single journey.
type RealmDefinition struct {
	Slug        string
	Name        string
	Order       int
	MaxProgress int
}

// RealmCatalog is the source of truth for journey progression metadata.
// Values are loaded from code-embedded defaults at startup and may be
// overridden by odyssey_system_config rows (key = "journey:<slug>:<field>").
type RealmCatalog struct {
	realms map[string]RealmDefinition
	order  []string
}

// ConfigLoader provides read access to system configuration values.
// It mirrors game.ConfigStore.GetSystemConfig to avoid an import cycle
// with pkg/game.
type ConfigLoader interface {
	GetSystemConfig(ctx context.Context, key string) (string, error)
}

type sysConfigRow struct {
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

func parseConfigValue(raw json.RawMessage) string {
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return s
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err == nil {
		return n.String()
	}
	return ""
}

// NewRealmCatalog builds a RealmCatalog from an ordered slice of definitions.
// The order of the slice determines the unlock sequence.
func NewRealmCatalog(defaults []RealmDefinition) *RealmCatalog {
	realms := make(map[string]RealmDefinition, len(defaults))
	order := make([]string, 0, len(defaults))
	for _, r := range defaults {
		realms[r.Slug] = r
		order = append(order, r.Slug)
	}
	return &RealmCatalog{realms: realms, order: order}
}

// Get returns the definition for a journey slug.
func (c *RealmCatalog) Get(slug string) (RealmDefinition, bool) {
	r, ok := c.realms[slug]
	return r, ok
}

// Order returns the journey unlock sequence (defensive copy).
func (c *RealmCatalog) Order() []string {
	out := make([]string, len(c.order))
	copy(out, c.order)
	return out
}

// NextRealm returns the journey that follows the given journey in the unlock
// sequence. Returns an empty string if the given journey is the last one or
// is unknown.
func (c *RealmCatalog) NextRealm(current string) string {
	for i, slug := range c.order {
		if slug == current && i+1 < len(c.order) {
			return c.order[i+1]
		}
	}
	return ""
}

// Override updates a single field on a journey definition. Supported fields:
// "name" (string), "max_progress" (positive integer). Returns false if the
// slug is unknown or the field/value is invalid.
func (c *RealmCatalog) Override(slug, field, value string) bool {
	r, ok := c.realms[slug]
	if !ok {
		return false
	}
	switch field {
	case "name":
		r.Name = value
		c.realms[slug] = r
		return true
	case "max_progress":
		var n int
		if _, err := fmt.Sscanf(value, "%d", &n); err == nil && n > 0 {
			r.MaxProgress = n
			c.realms[slug] = r
			return true
		}
		return false
	}
	return false
}

// ApplyOverrides reads journey-specific configuration from the provided ConfigLoader
// and applies them to matching journey definitions. The config key format is
// "journey:<slug>:<field>" where field is "name" or "max_progress".
// Rows that don't exist or have empty values are silently skipped.
func (c *RealmCatalog) ApplyOverrides(ctx context.Context, loader ConfigLoader) error {
	if loader == nil {
		return nil
	}
	const fieldsPerRealm = 2
	for _, slug := range c.order {
		for _, field := range []string{"name", "max_progress"} {
			key := fmt.Sprintf("journey:%s:%s", slug, field)
			raw, err := loader.GetSystemConfig(ctx, key)
			if err != nil {
				return fmt.Errorf("get system config %s: %w", key, err)
			}
			if raw == "" || raw == "[]" {
				continue
			}
			var rows []sysConfigRow
			if err := json.Unmarshal([]byte(raw), &rows); err != nil {
				return fmt.Errorf("parse system config %s: %w", key, err)
			}
			if len(rows) == 0 {
				continue
			}
			val := parseConfigValue(rows[0].Value)
			if val == "" {
				continue
			}
			c.Override(slug, field, val)
		}
	}
	return nil
}

// DefaultRealmDefinitions is the code-embedded journey metadata for the MVP.
// Each journey carries its slug, display name, unlock order, and maximum progress.
var DefaultRealmDefinitions = []RealmDefinition{
	{Slug: "whispering-woods", Name: "Whispering Woods", Order: 1, MaxProgress: 100},
	{Slug: "clockwork-city", Name: "Clockwork City", Order: 2, MaxProgress: 100},
	{Slug: "starlit-library", Name: "Starlit Library", Order: 3, MaxProgress: 100},
}

// DefaultRealmCatalog is the shared RealmCatalog instance initialised from
// DefaultRealmDefinitions. Tests may call NewRealmCatalog with custom
// definitions; production code uses this default.
var DefaultRealmCatalog = NewRealmCatalog(DefaultRealmDefinitions)
