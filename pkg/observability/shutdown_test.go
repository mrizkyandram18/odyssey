package observability

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"testing"
	"time"
)

func TestShutdownManager_Shutdown(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{Handler: handler}
	sm := NewShutdownManager(srv, 5*time.Second)

	called := false
	sm.AddHook("test", func(ctx context.Context) error {
		called = true
		return nil
	})

	err := sm.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("shutdown hook was not called")
	}
}

func TestShutdownManager_MultipleHooks(t *testing.T) {
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	sm := NewShutdownManager(srv, 5*time.Second)

	order := []int{}
	mu := &sync.Mutex{}

	for i := 0; i < 3; i++ {
		idx := i
		sm.AddHook("hook", func(ctx context.Context) error {
			mu.Lock()
			order = append(order, idx)
			mu.Unlock()
			return nil
		})
	}

	err := sm.Shutdown(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("expected 3 hooks called, got %d", len(order))
	}
}

func TestShutdownManager_HookError(t *testing.T) {
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	sm := NewShutdownManager(srv, 5*time.Second)

	sm.AddHook("ok1", func(ctx context.Context) error { return nil })
	sm.AddHook("fail", func(ctx context.Context) error { return errors.New("hook error") })
	sm.AddHook("ok2", func(ctx context.Context) error { return nil })

	err := sm.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected error from failing hook")
	}
}

func TestShutdownManager_RegisterServer(t *testing.T) {
	sm := NewShutdownManager(nil, 5*time.Second)
	srv := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})}
	sm.RegisterServer(srv)
	if sm.server != srv {
		t.Error("server not registered")
	}
}

func TestShutdownManager_DefaultTimeout(t *testing.T) {
	sm := NewShutdownManager(nil, 0)
	if sm.timeout != 30*time.Second {
		t.Errorf("expected default 30s timeout, got %v", sm.timeout)
	}
}
