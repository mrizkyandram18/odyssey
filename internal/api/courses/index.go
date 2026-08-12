package chapters

import (
	"context"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/course"
	"odyssey/pkg/shared"
)

type ChapterHandler interface {
	ListProgress(ctx context.Context, crewID string) ([]course.ChapterSummary, error)
	GetProgressView(ctx context.Context, crewID string) (*course.ChapterProgressView, error)
	Get(ctx context.Context, crewID, chapterSlug string) (*course.ChapterSummary, error)
	GetCurrentChapter(ctx context.Context, crewID string) (*course.ChapterSummary, error)
}

var handler ChapterHandler

func Setup(h ChapterHandler) {
	handler = h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims, ok := auth.ClaimsFromRequest(r)
	if !ok {
		shared.WriteUnauthorized(w)
		return
	}

	if handler == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	rest := strings.TrimPrefix(path, "api/courses")
	rest = strings.Trim(rest, "/")

	switch {
	case rest == "" || rest == "progress":
		handleProgress(w, r, claims)
	case rest == "current":
		handleCurrent(w, r, claims)
	default:
		handleGet(w, r, claims, rest)
	}
}

func handleProgress(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	view, err := handler.GetProgressView(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to get course progress", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, view)
}

func handleCurrent(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	result, err := handler.GetCurrentChapter(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to get current course", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}

func handleGet(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, slug string) {
	if !shared.ValidateSlug(slug) || len(slug) > 256 {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	result, err := handler.Get(r.Context(), claims.FamilyID, slug)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			shared.WriteJSONError(w, "course not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to get course", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}
