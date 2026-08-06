package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockContentProvider struct {
	err error
}

func (m *mockContentProvider) Status(ctx context.Context) (map[string]any, error) {
	if m.err != nil {
		return nil, m.err
	}
	return map[string]any{"ok": true}, nil
}

func (m *mockContentProvider) CacheHealthCheck() error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestVersionHandler(t *testing.T) {
	GitCommit = "abc123"
	BuildDate = "2024-01-15T10:30:00Z"
	Version = "1.2.3"
	SchemaVersion = "5"
	SetRuntimeSchemaVersion("")
	defer func() {
		GitCommit = "unknown"
		BuildDate = "unknown"
		Version = "dev"
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	bi := &StaticBuildInfo{ContentGen: 42}
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	VersionHandler(bi)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var vi VersionInfo
	if err := json.Unmarshal(w.Body.Bytes(), &vi); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if vi.GitCommit != "abc123" {
		t.Errorf("expected git_commit abc123, got %s", vi.GitCommit)
	}
	if vi.SemanticVersion != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %s", vi.SemanticVersion)
	}
	if vi.SchemaVersion != "5" {
		t.Errorf("expected schema_version 5, got %s", vi.SchemaVersion)
	}
	if vi.ContentGeneration != 42 {
		t.Errorf("expected content_generation 42, got %d", vi.ContentGeneration)
	}
}

func TestVersionHandler_NilBuildInfo(t *testing.T) {
	GitCommit = "test123"
	BuildDate = "unknown"
	Version = "1.0.0"
	SchemaVersion = "1"
	SetRuntimeSchemaVersion("")
	defer func() {
		GitCommit = "unknown"
		BuildDate = "unknown"
		Version = "dev"
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	VersionHandler(nil)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var vi VersionInfo
	if err := json.Unmarshal(w.Body.Bytes(), &vi); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if vi.GitCommit != "test123" {
		t.Errorf("expected git_commit test123, got %s", vi.GitCommit)
	}
	if vi.SemanticVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", vi.SemanticVersion)
	}
}

func TestVersionHandler_MethodNotAllowed(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/version", nil)
	w := httptest.NewRecorder()
	VersionHandler(nil)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestGetVersionInfo(t *testing.T) {
	GitCommit = "def456"
	BuildDate = "2024-06-01T12:00:00Z"
	Version = "2.0.0"
	SchemaVersion = "8"
	SetRuntimeSchemaVersion("")
	defer func() {
		GitCommit = "unknown"
		BuildDate = "unknown"
		Version = "dev"
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	vi := GetVersionInfo()
	if vi.GitCommit != "def456" {
		t.Errorf("expected GitCommit def456, got %s", vi.GitCommit)
	}
	if vi.SemanticVersion != "2.0.0" {
		t.Errorf("expected Version 2.0.0, got %s", vi.SemanticVersion)
	}
	if vi.SchemaVersion != "8" {
		t.Errorf("expected SchemaVersion 8, got %s", vi.SchemaVersion)
	}
}

func TestStaticBuildInfo(t *testing.T) {
	s := &StaticBuildInfo{ContentGen: 99}
	if s.GetContentGeneration() != 99 {
		t.Error("expected 99")
	}

	s = nil
	if s.GetContentGeneration() != 0 {
		t.Error("expected 0 for nil")
	}
}

func TestNoOpBuildInfo(t *testing.T) {
	s := NoOpBuildInfo()
	if s.GetContentGeneration() != 0 {
		t.Error("expected 0")
	}
}

func TestRuntimeSchemaVersion_OverridesBuildTime(t *testing.T) {
	SchemaVersion = "8"
	SetRuntimeSchemaVersion("")
	defer func() {
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	if SchemaVersionString() != "8" {
		t.Errorf("expected build-time fallback to 8, got %s", SchemaVersionString())
	}

	SetRuntimeSchemaVersion("12")
	if SchemaVersionString() != "12" {
		t.Errorf("expected runtime version 12, got %s", SchemaVersionString())
	}
}

func TestGetVersionInfo_UsesRuntimeSchemaVersion(t *testing.T) {
	GitCommit = "rt123"
	Version = "1.0.0"
	SchemaVersion = "8"
	SetRuntimeSchemaVersion("12")
	defer func() {
		GitCommit = "unknown"
		Version = "dev"
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	vi := GetVersionInfo()
	if vi.SchemaVersion != "12" {
		t.Errorf("expected runtime schema_version 12, got %s", vi.SchemaVersion)
	}
}

func TestReadSchemaVersion_Success(t *testing.T) {
	data := []byte(`[{"key":"schema_version","value":"12"}]`)
	client := &mockSupabaseClient{data: data}
	v, err := ReadSchemaVersion(context.Background(), client)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != "12" {
		t.Errorf("expected 12, got %s", v)
	}
}

func TestReadSchemaVersion_EmptyResult(t *testing.T) {
	data := []byte(`[]`)
	client := &mockSupabaseClient{data: data}
	_, err := ReadSchemaVersion(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for empty result")
	}
}

func TestReadSchemaVersion_DBError(t *testing.T) {
	client := &mockSupabaseClient{err: errDBUnavailable}
	_, err := ReadSchemaVersion(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for DB failure")
	}
}

func TestReadSchemaVersion_NilClient(t *testing.T) {
	_, err := ReadSchemaVersion(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil client")
	}
}
