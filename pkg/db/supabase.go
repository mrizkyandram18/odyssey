package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var allowedTables = map[string]bool{
	"odyssey_user_profiles":               true,
	"odyssey_crews":                       true,
	"odyssey_quests":                      true,
	"odyssey_challenges":                  true,
	"odyssey_realm_progress":              true,
	"odyssey_creative_items":              true,
	"odyssey_creative_submissions":        true,
	"odyssey_daily_turns":                 true,
	"odyssey_chests":                      true,
	"odyssey_relics":                      true,
	"odyssey_player_relics":               true,
	"odyssey_lore_unlocks":                true,
	"odyssey_chapter_progress":            true,
	"odyssey_achievements":                true,
	"odyssey_chest_definitions":           true,
	"odyssey_drop_tables":                 true,
	"odyssey_relic_definitions":           true,
	"odyssey_quest_definitions":           true,
	"odyssey_chapter_definitions":         true,
	"odyssey_realm_definitions":           true,
	"odyssey_creative_prompt_definitions": true,
	"odyssey_achievement_definitions":     true,
	"odyssey_season_definitions":          true,
	"odyssey_lore_definitions":            true,
	"odyssey_audit_logs":                  true,
	"odyssey_system_config":               true,
	"odyssey_balance_configs":             true,
	"odyssey_schema_version":              true,
}

func validateTable(table string) error {
	if table == "" {
		return fmt.Errorf("empty table name")
	}
	if !allowedTables[table] {
		return fmt.Errorf("table not allowed: %s", table)
	}
	return nil
}

type SupabaseClient interface {
	Get(ctx context.Context, table string, params string) ([]byte, error)
	Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error)
	MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error)
}

type supabaseClient struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

func NewClient(baseURL, apiKey string) SupabaseClient {
	return &supabaseClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  &http.Client{},
	}
}

func (c *supabaseClient) request(ctx context.Context, method, path, params string, body any) (*http.Response, error) {
	u := c.baseURL + restPath + path
	if params != "" {
		u += "?" + params
	}
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.client.Do(req)
}

const restPath = "/rest/v1/"

func (c *supabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if err := validateTable(table); err != nil {
		return nil, err
	}
	resp, err := c.request(ctx, http.MethodGet, table, params, nil)
	if err != nil {
		return nil, fmt.Errorf("supabase get: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase get %s: %s", resp.Status, raw)
	}
	return raw, nil
}

func (c *supabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return c.mutate(ctx, method, table, payload, params, "")
}

func (c *supabaseClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return c.mutate(ctx, method, table, payload, params, prefer)
}

func (c *supabaseClient) mutate(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	if err := validateTable(table); err != nil {
		return nil, err
	}
	u, err := url.Parse(c.baseURL + restPath + table)
	if err != nil {
		return nil, fmt.Errorf("parse url: %w", err)
	}
	if params != "" {
		q, err := url.ParseQuery(params)
		if err != nil {
			return nil, fmt.Errorf("parse params: %w", err)
		}
		u.RawQuery = q.Encode()
	}
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if prefer != "" {
		req.Header.Set("Prefer", prefer)
	}
	if strings.Contains(params, "return=representation") {
		req.Header.Set("Prefer", "return=representation")
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase mutate: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase mutate %s: %s", resp.Status, raw)
	}
	return raw, nil
}
