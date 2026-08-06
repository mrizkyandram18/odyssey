package world

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
)

type mockConfigLoader struct {
	configs map[string]string
}

func (m *mockConfigLoader) GetSystemConfig(ctx context.Context, key string) (string, error) {
	val, ok := m.configs[key]
	if !ok {
		return "[]", nil
	}
	if val == "" {
		return "[]", nil
	}
	row := sysConfigRow{Key: key, Value: json.RawMessage(fmt.Sprintf("%q", val))}
	out, _ := json.Marshal([]sysConfigRow{row})
	return string(out), nil
}

func TestApplyOverrides_MaxProgress(t *testing.T) {
	catalog := NewRealmCatalog(DefaultRealmDefinitions)
	loader := &mockConfigLoader{
		configs: map[string]string{
			"realm:whispering-woods:max_progress": "200",
			"realm:clockwork-city:max_progress":   "150",
		},
	}

	if err := catalog.ApplyOverrides(context.Background(), loader); err != nil {
		t.Fatalf("ApplyOverrides failed: %v", err)
	}

	def, ok := catalog.Get("whispering-woods")
	if !ok {
		t.Fatal("whispering-woods not found")
	}
	if def.MaxProgress != 200 {
		t.Errorf("whispering-woods MaxProgress = %d, want 200", def.MaxProgress)
	}

	def, ok = catalog.Get("clockwork-city")
	if !ok {
		t.Fatal("clockwork-city not found")
	}
	if def.MaxProgress != 150 {
		t.Errorf("clockwork-city MaxProgress = %d, want 150", def.MaxProgress)
	}

	def, ok = catalog.Get("starlit-library")
	if !ok {
		t.Fatal("starlit-library not found")
	}
	if def.MaxProgress != 100 {
		t.Errorf("starlit-library MaxProgress = %d, want 100 (unchanged)", def.MaxProgress)
	}
}

func TestApplyOverrides_Name(t *testing.T) {
	catalog := NewRealmCatalog(DefaultRealmDefinitions)
	loader := &mockConfigLoader{
		configs: map[string]string{
			"realm:whispering-woods:name": "Whispering Woods (Hard Mode)",
		},
	}

	if err := catalog.ApplyOverrides(context.Background(), loader); err != nil {
		t.Fatalf("ApplyOverrides failed: %v", err)
	}

	def, ok := catalog.Get("whispering-woods")
	if !ok {
		t.Fatal("realm not found")
	}
	if def.Name != "Whispering Woods (Hard Mode)" {
		t.Errorf("Name = %q, want %q", def.Name, "Whispering Woods (Hard Mode)")
	}
}

func TestApplyOverrides_NilLoader(t *testing.T) {
	catalog := NewRealmCatalog(DefaultRealmDefinitions)
	if err := catalog.ApplyOverrides(context.Background(), nil); err != nil {
		t.Fatalf("ApplyOverrides with nil loader should not error: %v", err)
	}
}

func TestApplyOverrides_NoOverrides(t *testing.T) {
	catalog := NewRealmCatalog(DefaultRealmDefinitions)
	loader := &mockConfigLoader{
		configs: map[string]string{},
	}

	if err := catalog.ApplyOverrides(context.Background(), loader); err != nil {
		t.Fatalf("ApplyOverrides failed: %v", err)
	}

	def, ok := catalog.Get("whispering-woods")
	if !ok || def.MaxProgress != 100 {
		t.Errorf("MaxProgress should remain 100, got %d", def.MaxProgress)
	}
}

func TestApplyOverrides_InvalidMaxProgress(t *testing.T) {
	catalog := NewRealmCatalog(DefaultRealmDefinitions)
	loader := &mockConfigLoader{
		configs: map[string]string{
			"realm:whispering-woods:max_progress": "abc",
		},
	}

	if err := catalog.ApplyOverrides(context.Background(), loader); err != nil {
		t.Fatalf("ApplyOverrides should not error on invalid override: %v", err)
	}

	def, _ := catalog.Get("whispering-woods")
	if def.MaxProgress != 100 {
		t.Errorf("Invalid max_progress should leave default 100, got %d", def.MaxProgress)
	}
}
