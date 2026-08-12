package story_fragments

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/fragment"
	"odyssey/pkg/shared"
)

type FragmentHandler interface {
	ListPlayerFragments(ctx context.Context, uid string) ([]fragment.StoryFragmentView, error)
	DiscoverFragment(ctx context.Context, uid, crewID, slug string) (*fragment.DiscoverResult, error)
	ReplayRealm(ctx context.Context, uid, crewID, journey string) (*fragment.ReplayResult, error)
}

var handler FragmentHandler

func Setup(h FragmentHandler) {
	handler = h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

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
	rest := strings.TrimPrefix(path, "api/story_fragments")
	rest = strings.Trim(rest, "/")

	switch {
	case r.Method == http.MethodGet && rest == "":
		handleList(w, r, claims)
	case r.Method == http.MethodPost && rest == "discover":
		handleDiscover(w, r, claims)
	case r.Method == http.MethodPost && rest == "replay":
		handleReplay(w, r, claims)
	default:
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
	}
}

func handleList(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	frags, err := handler.ListPlayerFragments(r.Context(), claims.UID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list story fragments", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, frags)
}

type discoverReq struct {
	Slug string `json:"slug"`
}

func handleDiscover(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	var req discoverReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Slug == "" {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := handler.DiscoverFragment(r.Context(), claims.UID, claims.FamilyID, req.Slug)
	if err != nil {
		if errors.Is(err, fragment.ErrFragmentNotFound) {
			shared.WriteJSONError(w, "story fragment not found", http.StatusNotFound)
			return
		}
		if errors.Is(err, fragment.ErrUnauthorized) {
			shared.WriteUnauthorized(w)
			return
		}
		shared.WriteJSONError(w, "failed to discover fragment", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, result)
}

type replayReq struct {
	Journey string `json:"journey"`
}

func handleReplay(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	var req replayReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Journey == "" {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := handler.ReplayRealm(r.Context(), claims.UID, claims.FamilyID, req.Journey)
	if err != nil {
		if errors.Is(err, fragment.ErrUnauthorized) {
			shared.WriteUnauthorized(w)
			return
		}
		shared.WriteJSONError(w, "failed to replay journey", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, result)
}
