package shared

import (
	"fmt"
	"net/http"
	"strings"
)

func ValidateSlug(slug string) bool {
	if slug == "" {
		return false
	}
	if len(slug) > 256 {
		return false
	}
	for _, c := range slug {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_') {
			return false
		}
	}
	return true
}

func ValidateInt64Param(r *http.Request, key string) (int64, bool) {
	val := r.URL.Query().Get(key)
	if val == "" {
		return 0, false
	}
	var n int64
	_, err := fmt.Sscanf(val, "%d", &n)
	return n, err == nil
}

func SanitizeString(s string, maxLen int) string {
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	s = strings.TrimSpace(s)
	return s
}
