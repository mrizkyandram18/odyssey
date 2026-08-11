package quests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/quest"
	"odyssey/pkg/shared"
)

// QuestHandler is the interface the handler depends on. The concrete
// implementation (pkg/game/quest.QuestAPIHandler) composes QuestService with
// the progression and realm-progress collaborators.
type QuestHandler interface {
	List(ctx context.Context, crewID string) ([]quest.QuestView, error)
	ListAvailable(ctx context.Context, crewID, uid string) ([]quest.QuestView, error)
	GetByCrewAndID(ctx context.Context, questID int64, crewID string) (*quest.QuestWithChallenges, error)
	StartQuest(ctx context.Context, questID int64, crewID string) error
	CompleteChallenge(ctx context.Context, questID, challengeID int64, crewID, uid string) (*quest.CompleteChallengeResult, error)
	SelectBranch(ctx context.Context, questID int64, crewID, branchChoice string) (*quest.SelectBranchResult, error)
}

var handler QuestHandler

func Setup(h QuestHandler) {
	handler = h
}

// routeInfo captures the parsed quest resource path:
//
//	/api/quests                      → list
//	/api/quests/<id>                 → get by id
//	/api/quests/<id>/start           → start quest (POST)
//	/api/quests/<id>/branch          → select narrative branch (POST)
//	/api/quests/<id>/challenges/<cid>/complete → complete challenge (POST)
type routeInfo struct {
	idStr  string // raw quest id segment ("" when absent)
	action string // "start", "branch", "complete", or ""
	cidStr string // raw challenge id segment ("" when absent)
}

func parseQuestPath(path string) routeInfo {
	trimmed := strings.Trim(path, "/")
	if !strings.HasPrefix(trimmed, "api/quests") {
		return routeInfo{}
	}
	rest := strings.TrimPrefix(trimmed, "api/quests")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return routeInfo{}
	}

	parts := strings.Split(rest, "/")
	switch {
	case len(parts) == 1:
		return routeInfo{idStr: parts[0]}
	case len(parts) == 2 && parts[1] == "start":
		return routeInfo{idStr: parts[0], action: "start"}
	case len(parts) == 2 && parts[1] == "branch":
		return routeInfo{idStr: parts[0], action: "branch"}
	case len(parts) == 4 && parts[1] == "challenges" && parts[3] == "complete":
		return routeInfo{idStr: parts[0], action: "complete", cidStr: parts[2]}
	}
	return routeInfo{}
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	route := parseQuestPath(r.URL.Path)

	// Route selection happens before auth so that unsupported methods/paths
	// receive a 405 regardless of authentication state.
	var routeFn func(claims *auth.SessionClaims)
	switch {
	case r.Method == http.MethodGet && route.idStr == "" && route.action == "":
		routeFn = func(c *auth.SessionClaims) { handleList(w, r, c) }
	case r.Method == http.MethodGet && route.idStr == "available" && route.action == "":
		routeFn = func(c *auth.SessionClaims) { handleListAvailable(w, r, c) }
	case r.Method == http.MethodGet && route.action == "" && route.idStr != "" && route.cidStr == "":
		idStr := route.idStr
		routeFn = func(c *auth.SessionClaims) { handleGetByID(w, r, c, idStr) }
	case r.Method == http.MethodPost && route.action == "start":
		idStr := route.idStr
		routeFn = func(c *auth.SessionClaims) { handleStart(w, r, c, idStr) }
	case r.Method == http.MethodPost && route.action == "branch":
		idStr := route.idStr
		routeFn = func(c *auth.SessionClaims) { handleSelectBranch(w, r, c, idStr) }
	case r.Method == http.MethodPost && route.action == "complete":
		idStr, cidStr := route.idStr, route.cidStr
		routeFn = func(c *auth.SessionClaims) { handleCompleteChallenge(w, r, c, idStr, cidStr) }
	default:
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

	routeFn(claims)
}

func handleList(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	quests, err := handler.List(r.Context(), claims.CrewID)
	if err != nil {
		shared.WriteJSONError(w, "failed to load quests", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, quests)
}

func handleListAvailable(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	quests, err := handler.ListAvailable(r.Context(), claims.CrewID, claims.UID)
	if err != nil {
		shared.WriteJSONError(w, "failed to load available quests", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, quests)
}

func handleGetByID(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr string) {
	questID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid quest id", http.StatusBadRequest)
		return
	}

	q, err := handler.GetByCrewAndID(r.Context(), questID, claims.CrewID)
	if err != nil {
		if isNotFound(err) {
			shared.WriteJSONError(w, "quest not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to load quest", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, q)
}

func handleStart(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr string) {
	questID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid quest id", http.StatusBadRequest)
		return
	}

	if err := handler.StartQuest(r.Context(), questID, claims.CrewID); err != nil {
		if isNotFound(err) {
			shared.WriteJSONError(w, "quest not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to start quest", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]bool{"started": true})
}

func handleCompleteChallenge(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr, cidStr string) {
	questID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid quest id", http.StatusBadRequest)
		return
	}
	challengeID, err := strconv.ParseInt(cidStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid challenge id", http.StatusBadRequest)
		return
	}

	result, err := handler.CompleteChallenge(r.Context(), questID, challengeID, claims.CrewID, claims.UID)
	if err != nil {
		if isNotFound(err) {
			shared.WriteJSONError(w, "quest not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to complete challenge", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, result)
}

func isNotFound(err error) bool {
	return strings.Contains(err.Error(), "not found")
}

type selectBranchReq struct {
	Branch string `json:"branch"`
}

func handleSelectBranch(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr string) {
	questID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid quest id", http.StatusBadRequest)
		return
	}

	var req selectBranchReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Branch == "" {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	result, err := handler.SelectBranch(r.Context(), questID, claims.CrewID, req.Branch)
	if err != nil {
		if errors.Is(err, quest.ErrInvalidBranchChoice) || errors.Is(err, quest.ErrNoBranchOptions) {
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		if isNotFound(err) || errors.Is(err, quest.ErrQuestNotFound) {
			shared.WriteJSONError(w, "quest not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to select branch", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, result)
}
