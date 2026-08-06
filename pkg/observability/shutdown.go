package observability

import (
	"context"
	"net/http"
	"sync"
	"time"
)

type ShutdownHook func(ctx context.Context) error

type ShutdownManager struct {
	server  *http.Server
	hooks   []ShutdownHook
	timeout time.Duration
}

func NewShutdownManager(server *http.Server, timeout time.Duration) *ShutdownManager {
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	return &ShutdownManager{
		server:  server,
		timeout: timeout,
	}
}

func (sm *ShutdownManager) AddHook(name string, hook ShutdownHook) {
	sm.hooks = append(sm.hooks, hook)
}

func (sm *ShutdownManager) Shutdown(ctx context.Context) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), sm.timeout)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i := range sm.hooks {
		hook := sm.hooks[i]
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := hook(shutdownCtx); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
			}
		}()
	}

	_ = sm.server.Shutdown(shutdownCtx)
	wg.Wait()

	return firstErr
}

func (sm *ShutdownManager) RegisterServer(server *http.Server) {
	sm.server = server
}
