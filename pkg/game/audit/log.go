package audit

import (
	"context"
	"encoding/json"
	"time"

	"odyssey/pkg/observability"
)

// Operation types for audit logging.
const (
	OpCreate    = "create"
	OpUpdate    = "update"
	OpDelete    = "delete"
	OpRestore   = "restore"
	OpPublish   = "publish"
	OpSaveDraft = "save_draft"
	OpReload    = "reload"
	OpValidate  = "validate"
)

// AuditEntry records a single admin mutation against a content resource.
type AuditEntry struct {
	ID         int64           `json:"id"`
	Resource   string          `json:"resource"`
	ResourceID string          `json:"resource_id,omitempty"`
	Operation  string          `json:"operation"`
	AdminUID   string          `json:"admin_uid"`
	RequestID  string          `json:"request_id,omitempty"`
	OldValue   json.RawMessage `json:"old_value,omitempty"`
	NewValue   json.RawMessage `json:"new_value,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// Store provides persistence for audit log entries.
type Store interface {
	Write(ctx context.Context, entry AuditEntry) error
	List(ctx context.Context, filter ListFilter) ([]AuditEntry, error)
}

// ListFilter restricts audit log queries.
type ListFilter struct {
	Resource  string
	AdminUID  string
	Operation string
	Limit     int
}

// Logger wraps a Store and provides convenient Write methods that
// capture old/new values as JSON.
type Logger struct {
	store Store
}

func NewLogger(store Store) *Logger {
	return &Logger{store: store}
}

func (l *Logger) Log(ctx context.Context, resource, resourceID, operation, adminUID string, oldValue, newValue any) error {
	requestID := observability.RequestIDFromContext(ctx)
	oldJSON, _ := json.Marshal(oldValue)
	newJSON, _ := json.Marshal(newValue)
	entry := AuditEntry{
		Resource:   resource,
		ResourceID: resourceID,
		Operation:  operation,
		AdminUID:   adminUID,
		RequestID:  requestID,
		OldValue:   json.RawMessage(oldJSON),
		NewValue:   json.RawMessage(newJSON),
		CreatedAt:  time.Now().UTC(),
	}
	return l.store.Write(ctx, entry)
}

func (l *Logger) LogError(ctx context.Context, resource, resourceID, operation, adminUID string, err error) {
	_ = l.Log(ctx, resource, resourceID, operation, adminUID, nil, map[string]string{"error": err.Error()})
}
