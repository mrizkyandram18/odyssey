package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	id := generateRequestID()
	if len(id) == 0 {
		t.Fatal("expected non-empty request ID")
	}
	if id == generateRequestID() {
		// Extremely unlikely collision for 128-bit random, but check format
		t.Fatal("expected unique request IDs")
	}
}

func TestRequestIDMiddleware_GeneratesNewID(t *testing.T) {
	called := false
	handler := RequestIDMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		id := r.Context().Value(requestIDKey)
		if id == nil {
			t.Fatal("expected request ID in context")
		}
		if id.(string) != r.Header.Get(HeaderRequestID) && r.Header.Get(HeaderRequestID) != "" {
			t.Fatal("context request ID does not match incoming header")
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if w.Header().Get(HeaderRequestID) == "" {
		t.Fatal("expected X-Request-ID response header to be set")
	}
}

func TestRequestIDMiddleware_PreservesExistingID(t *testing.T) {
	called := false
	existingID := "test-request-id-123"
	handler := RequestIDMiddleware(func(w http.ResponseWriter, r *http.Request) {
		called = true
		id := r.Context().Value(requestIDKey)
		if id == nil {
			t.Fatal("expected request ID in context")
		}
		if id.(string) != existingID {
			t.Fatalf("expected %s, got %s", existingID, id)
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(HeaderRequestID, existingID)
	w := httptest.NewRecorder()
	handler(w, req)

	if !called {
		t.Fatal("handler was not called")
	}
	if w.Header().Get(HeaderRequestID) != existingID {
		t.Fatalf("expected header %s, got %s", existingID, w.Header().Get(HeaderRequestID))
	}
}

func TestRequestIDFromContext(t *testing.T) {
	id := "test-ctx-id"
	ctx := WithRequestID(context.Background(), id)
	got := RequestIDFromContext(ctx)
	if got != id {
		t.Fatalf("expected %s, got %s", id, got)
	}
}

func TestRequestIDFromContext_Missing(t *testing.T) {
	got := RequestIDFromContext(context.Background())
	if got != "" {
		t.Fatalf("expected empty string, got %s", got)
	}
}

func TestRequestIDMiddleware_PropagatesToHandler(t *testing.T) {
	var capturedID string
	handler := RequestIDMiddleware(func(w http.ResponseWriter, r *http.Request) {
		capturedID = RequestIDFromContext(r.Context())
	})

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if capturedID == "" {
		t.Fatal("expected request ID to be captured in handler")
	}
}
