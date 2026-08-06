package shared

import (
	"encoding/json"
	"errors"
	"net/http"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrBadRequest   = errors.New("bad request")
	ErrInternal     = errors.New("internal error")
)

func ReadJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func WriteJSONError(w http.ResponseWriter, msg string, status int) {
	WriteJSON(w, status, map[string]string{"error": msg})
}

func WriteUnauthorized(w http.ResponseWriter) {
	WriteJSONError(w, "unauthorized", http.StatusUnauthorized)
}

func WriteForbidden(w http.ResponseWriter) {
	WriteJSONError(w, "forbidden", http.StatusForbidden)
}
