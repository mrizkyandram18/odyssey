package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var allowedTables = map[string]bool{
	"odyssey_user_profiles":       true,
	"odyssey_families":            true,
	"odyssey_local_users":         true,
	"odyssey_tasks":               true,
	"odyssey_task_submissions":    true,
	"odyssey_reward_catalog":      true,
	"odyssey_claims":              true,
	"odyssey_coin_transactions":   true,
	"odyssey_push_subscriptions":  true,
	"odyssey_schema_version":      true,
	"odyssey_system_config":       true,
	"odyssey_user_payout_config":  true,
	"odyssey_member_monthly_targets": true,
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
	RPC(ctx context.Context, fnName string, payload any) ([]byte, error)
	UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error)
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
		if q.Has("return") {
			if prefer == "" {
				prefer = "return=" + q.Get("return")
			}
			q.Del("return")
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

func (c *supabaseClient) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	u := c.baseURL + restPath + "rpc/" + fnName
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("marshal rpc payload: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bodyReader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("supabase rpc %s: %w", fnName, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rpc response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase rpc %s %s: %s", fnName, resp.Status, raw)
	}
	return raw, nil
}

func (c *supabaseClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	u := fmt.Sprintf("%s/storage/v1/object/%s/%s", c.baseURL, bucket, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("apikey", c.apiKey)
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	} else {
		req.Header.Set("Content-Type", "application/octet-stream")
	}
	req.Header.Set("x-upsert", "true")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("storage upload error: %w", err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("storage upload %s: %s", resp.Status, raw)
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", c.baseURL, bucket, path)
	return publicURL, nil
}
