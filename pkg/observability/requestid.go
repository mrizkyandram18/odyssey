package observability

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

func generateRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "req-" + randomHex(16)
	}
	return hex.EncodeToString(b[:])
}

func randomHex(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = hexDigits[i%16]
	}
	return string(b)
}

const hexDigits = "0123456789abcdef"

func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = generateRequestID()
		}
		ctx := WithRequestID(r.Context(), requestID)
		w.Header().Set(HeaderRequestID, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
