package observability

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

var (
	GitCommit     = "unknown"
	BuildDate     = "unknown"
	Version       = "dev"
	SchemaVersion = "0"
)

var (
	runtimeSchemaVersion string
	runtimeSchemaMu      sync.RWMutex
)

// SetRuntimeSchemaVersion overrides the build-time SchemaVersion with the
// value read from the database at startup. This is the source of truth
// at runtime — the build-time variable only records what the binary
// *expects*.
func SetRuntimeSchemaVersion(v string) {
	runtimeSchemaMu.Lock()
	runtimeSchemaVersion = v
	runtimeSchemaMu.Unlock()
}

// SchemaVersionString returns the runtime schema version if it was set
// (via SetRuntimeSchemaVersion), otherwise falls back to the build-time
// SchemaVersion variable.
func SchemaVersionString() string {
	runtimeSchemaMu.RLock()
	v := runtimeSchemaVersion
	runtimeSchemaMu.RUnlock()
	if v != "" {
		return v
	}
	return SchemaVersion
}

// ReadSchemaVersion reads the current schema version from the
// odyssey_schema_version table in the database. Returns the build-time
// SchemaVersion (or "0") if the database is unavailable so that the
// server can still start in degraded mode.
func ReadSchemaVersion(ctx context.Context, client SupabaseClient) (string, error) {
	if client == nil {
		return "", fmt.Errorf("nil supabase client")
	}
	raw, err := client.Get(ctx, "odyssey_schema_version", "key=eq.schema_version")
	if err != nil {
		return "", fmt.Errorf("read schema version: %w", err)
	}
	var rows []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return "", fmt.Errorf("parse schema version: %w", err)
	}
	if len(rows) == 0 {
		return "", fmt.Errorf("schema version row not found")
	}
	return rows[0].Value, nil
}

type VersionInfo struct {
	GitCommit         string    `json:"git_commit"`
	BuildDate         string    `json:"build_date"`
	SemanticVersion   string    `json:"semantic_version"`
	SchemaVersion     string    `json:"schema_version"`
	ContentGeneration int64     `json:"content_generation,omitempty"`
	BuildTime         time.Time `json:"build_time,omitempty"`
}

func GetVersionInfo() VersionInfo {
	var bt time.Time
	if BuildDate != "unknown" {
		if t, err := time.Parse(time.RFC3339, BuildDate); err == nil {
			bt = t
		}
	}
	return VersionInfo{
		GitCommit:       GitCommit,
		BuildDate:       BuildDate,
		SemanticVersion: Version,
		SchemaVersion:   SchemaVersionString(),
		BuildTime:       bt,
	}
}

type BuildInfoProvider interface {
	GetContentGeneration() int64
}
