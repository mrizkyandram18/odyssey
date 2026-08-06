package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"odyssey/pkg/game/content"
)

type supabaseDefinitionStore struct {
	client SupabaseClient
}

func NewDefinitionStore(client SupabaseClient) content.AdminStore {
	return &supabaseDefinitionStore{client: client}
}

// withPublishedFilter adds published=eq.true and deleted_at=is.null to params.
func withPublishedFilter(v url.Values) {
	v.Set("published", "eq.true")
	v.Set("deleted_at", "is.null")
}

func (s *supabaseDefinitionStore) GetByID(ctx context.Context, table string, id int64) (map[string]any, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(id, 10))
	raw, err := s.client.Get(ctx, table, v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get by id from %s: %w", table, err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse response from %s: %w", table, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func (s *supabaseDefinitionStore) GetBySlug(ctx context.Context, table, slug string) (map[string]any, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	raw, err := s.client.Get(ctx, table, v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get by slug from %s: %w", table, err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse response from %s: %w", table, err)
	}
	if len(rows) == 0 {
		return nil, nil
	}
	return rows[0], nil
}

func (s *supabaseDefinitionStore) ListAll(ctx context.Context, table string, includeDeleted bool) ([]map[string]any, error) {
	v := url.Values{}
	v.Set("order", "id.desc")
	if !includeDeleted {
		withPublishedFilter(v)
	} else {
		v.Set("deleted_at", "is.null")
	}
	raw, err := s.client.Get(ctx, table, v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list all from %s: %w", table, err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse response from %s: %w", table, err)
	}
	return rows, nil
}

func (s *supabaseDefinitionStore) ListPublished(ctx context.Context, table string) ([]map[string]any, error) {
	raw, err := s.client.Get(ctx, table, publishedParams())
	if err != nil {
		return nil, fmt.Errorf("list published from %s: %w", table, err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse response from %s: %w", table, err)
	}
	return rows, nil
}

func (s *supabaseDefinitionStore) Create(ctx context.Context, table string, data map[string]any) (map[string]any, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	data["published"] = false
	data["version"] = 1
	if _, ok := data["created_at"]; !ok {
		data["created_at"] = now
	}
	data["updated_at"] = now

	raw, err := s.client.Mutate(ctx, "POST", table, data, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create in %s: %w", table, err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse created row from %s: %w", table, err)
	}
	if len(rows) == 0 {
		return data, nil
	}
	return rows[0], nil
}

func (s *supabaseDefinitionStore) UpdateDraft(ctx context.Context, table, slug string, patchData map[string]any, updatedBy string) error {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	raw, err := s.client.Get(ctx, table, v.Encode())
	if err != nil {
		return fmt.Errorf("get current version for draft update: %w", err)
	}
	var rows []map[string]any
	if err := json.Unmarshal(raw, &rows); err != nil {
		return fmt.Errorf("parse current version: %w", err)
	}
	if len(rows) == 0 {
		return fmt.Errorf("definition not found in %s with slug %s", table, slug)
	}
	currentVersion := 0
	if ver, ok := rows[0]["version"].(float64); ok {
		currentVersion = int(ver)
	}

	patchData["version"] = currentVersion + 1
	patchData["updated_by"] = updatedBy
	patchData["updated_at"] = time.Now().UTC().Format(time.RFC3339)

	params := v.Encode()
	_, err = s.client.Mutate(ctx, "PATCH", table, patchData, params)
	if err != nil {
		return fmt.Errorf("update draft in %s: %w", table, err)
	}
	return nil
}

func (s *supabaseDefinitionStore) Publish(ctx context.Context, table, slug string, updatedBy string) error {
	raw, err := s.GetBySlug(ctx, table, slug)
	if err != nil {
		return fmt.Errorf("publish get draft: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("definition not found in %s with slug %s", table, slug)
	}

	patch := map[string]any{
		"published":    true,
		"published_at": time.Now().UTC().Format(time.RFC3339),
		"updated_by":   updatedBy,
		"updated_at":   time.Now().UTC().Format(time.RFC3339),
	}

	currentVersion := 0
	if ver, ok := raw["version"].(float64); ok {
		currentVersion = int(ver)
	}
	patch["version"] = currentVersion + 1

	v := url.Values{}
	v.Set("slug", "eq."+slug)
	params := v.Encode()
	_, err = s.client.Mutate(ctx, "PATCH", table, patch, params)
	if err != nil {
		return fmt.Errorf("publish in %s: %w", table, err)
	}
	return nil
}

func (s *supabaseDefinitionStore) SoftDelete(ctx context.Context, table, slug string) error {
	raw, err := s.GetBySlug(ctx, table, slug)
	if err != nil {
		return fmt.Errorf("soft delete get: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("definition not found in %s with slug %s", table, slug)
	}
	currentVersion := 0
	if ver, ok := raw["version"].(float64); ok {
		currentVersion = int(ver)
	}

	patch := map[string]any{
		"deleted_at": time.Now().UTC().Format(time.RFC3339),
		"published":  false,
		"version":    currentVersion + 1,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	v := url.Values{}
	v.Set("slug", "eq."+slug)
	params := v.Encode()
	_, err = s.client.Mutate(ctx, "PATCH", table, patch, params)
	if err != nil {
		return fmt.Errorf("soft delete in %s: %w", table, err)
	}
	return nil
}

func (s *supabaseDefinitionStore) Restore(ctx context.Context, table, slug string) error {
	raw, err := s.GetBySlug(ctx, table, slug)
	if err != nil {
		return fmt.Errorf("restore get: %w", err)
	}
	if raw == nil {
		return fmt.Errorf("definition not found in %s with slug %s", table, slug)
	}
	currentVersion := 0
	if ver, ok := raw["version"].(float64); ok {
		currentVersion = int(ver)
	}

	wasPublished := true
	if pub, ok := raw["published"].(bool); ok {
		wasPublished = pub
	}

	patch := map[string]any{
		"deleted_at": nil,
		"published":  wasPublished,
		"version":    currentVersion + 1,
		"updated_at": time.Now().UTC().Format(time.RFC3339),
	}

	v := url.Values{}
	v.Set("slug", "eq."+slug)
	params := v.Encode()
	_, err = s.client.Mutate(ctx, "PATCH", table, patch, params)
	if err != nil {
		return fmt.Errorf("restore in %s: %w", table, err)
	}
	return nil
}

var _ content.AdminStore = (*supabaseDefinitionStore)(nil)
