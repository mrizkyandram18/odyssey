package observability

import (
	"context"
	"time"
)

type SupabaseClient interface {
	Get(ctx context.Context, table string, params string) ([]byte, error)
	Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error)
	MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error)
}

type ProfilingClient struct {
	inner   SupabaseClient
	metrics *Metrics
}

func NewProfilingClient(inner SupabaseClient, metrics *Metrics) *ProfilingClient {
	return &ProfilingClient{inner: inner, metrics: metrics}
}

func (c *ProfilingClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	start := time.Now()
	result, err := c.inner.Get(ctx, table, params)
	dur := time.Since(start)
	c.record(ctx, table, "GET", params, dur)
	return result, err
}

func (c *ProfilingClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	start := time.Now()
	result, err := c.inner.Mutate(ctx, method, table, payload, params)
	dur := time.Since(start)
	c.record(ctx, table, method, params, dur)
	return result, err
}

func (c *ProfilingClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	start := time.Now()
	result, err := c.inner.MutateAtomic(ctx, method, table, payload, params, prefer)
	dur := time.Since(start)
	c.record(ctx, table, method, params, dur)
	return result, err
}

func (c *ProfilingClient) record(ctx context.Context, table, method, params string, dur time.Duration) {
	if c.metrics != nil {
		c.metrics.RecordDBLatency(dur)
	}
	if rp := profileFromContext(ctx); rp != nil {
		rp.RecordDBQuery(table, method, params, dur)
	}
}

var _ SupabaseClient = (*ProfilingClient)(nil)
