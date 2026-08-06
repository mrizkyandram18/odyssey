package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"odyssey/pkg/game/audit"
)

type AuditEntry struct {
	ID         int64           `json:"id"`
	Resource   string          `json:"resource"`
	ResourceID string          `json:"resource_id"`
	Operation  string          `json:"operation"`
	AdminUID   string          `json:"admin_uid"`
	RequestID  string          `json:"request_id,omitempty"`
	OldValue   json.RawMessage `json:"old_value"`
	NewValue   json.RawMessage `json:"new_value"`
	CreatedAt  time.Time       `json:"created_at"`
}

type supabaseAuditStore struct {
	client SupabaseClient
}

func NewAuditStore(client SupabaseClient) audit.Store {
	return &supabaseAuditStore{client: client}
}

func (s *supabaseAuditStore) Write(ctx context.Context, entry audit.AuditEntry) error {
	dbEntry := AuditEntry{
		Resource:   entry.Resource,
		ResourceID: entry.ResourceID,
		Operation:  entry.Operation,
		AdminUID:   entry.AdminUID,
		RequestID:  entry.RequestID,
		OldValue:   entry.OldValue,
		NewValue:   entry.NewValue,
	}
	payload := map[string]any{
		"resource":    dbEntry.Resource,
		"resource_id": dbEntry.ResourceID,
		"operation":   dbEntry.Operation,
		"admin_uid":   dbEntry.AdminUID,
	}
	if dbEntry.RequestID != "" {
		payload["request_id"] = dbEntry.RequestID
	}
	if len(dbEntry.OldValue) > 0 {
		var oldVal any
		if err := json.Unmarshal(dbEntry.OldValue, &oldVal); err == nil {
			payload["old_value"] = oldVal
		}
	}
	if len(dbEntry.NewValue) > 0 {
		var newVal any
		if err := json.Unmarshal(dbEntry.NewValue, &newVal); err == nil {
			payload["new_value"] = newVal
		}
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_audit_logs", payload, "")
	if err != nil {
		return fmt.Errorf("write audit log: %w", err)
	}
	return nil
}

func (s *supabaseAuditStore) List(ctx context.Context, filter audit.ListFilter) ([]audit.AuditEntry, error) {
	v := url.Values{}
	if filter.Resource != "" {
		v.Set("resource", "eq."+filter.Resource)
	}
	if filter.AdminUID != "" {
		v.Set("admin_uid", "eq."+filter.AdminUID)
	}
	if filter.Operation != "" {
		v.Set("operation", "eq."+filter.Operation)
	}
	v.Set("order", "created_at.desc")
	if filter.Limit > 0 {
		v.Set("limit", strconv.Itoa(filter.Limit))
	}
	raw, err := s.client.Get(ctx, "odyssey_audit_logs", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	var rows []AuditEntry
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse audit logs: %w", err)
	}
	result := make([]audit.AuditEntry, 0, len(rows))
	for i := range rows {
		result = append(result, audit.AuditEntry{
			ID:         rows[i].ID,
			Resource:   rows[i].Resource,
			ResourceID: rows[i].ResourceID,
			Operation:  rows[i].Operation,
			AdminUID:   rows[i].AdminUID,
			OldValue:   rows[i].OldValue,
			NewValue:   rows[i].NewValue,
			CreatedAt:  rows[i].CreatedAt,
		})
	}
	return result, nil
}

var _ audit.Store = (*supabaseAuditStore)(nil)
