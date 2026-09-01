package shared

import (
	"net/http"
	"strconv"
)

type PaginationMeta struct {
	Page    int  `json:"page"`
	Limit   int  `json:"limit"`
	Total   int  `json:"total"`
	HasNext bool `json:"has_next"`
}

type PaginatedResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationMeta `json:"pagination"`
}

// ParsePagination extracts and clamps page and limit query parameters safely.
// Default limit is used if limit param is missing or <= 0.
// Max limit is enforced (default max: 100).
func ParsePagination(r *http.Request, defaultLimit, maxLimit int) (int, int, int) {
	if defaultLimit <= 0 {
		defaultLimit = 50
	}
	if maxLimit <= 0 {
		maxLimit = 100
	}

	page := 1
	if pStr := r.URL.Query().Get("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p >= 1 {
			page = p
		}
	}

	limit := defaultLimit
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l >= 1 {
			if l > maxLimit {
				limit = maxLimit
			} else {
				limit = l
			}
		}
	}

	offset := (page - 1) * limit
	return page, limit, offset
}
