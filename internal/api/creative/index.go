package creative

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/game/creative"
	"odyssey/pkg/shared"
)

// CreativeHandler is the interface the handler depends on.
type CreativeHandler interface {
	Submit(ctx context.Context, uid string, req *game.Submission) (*creative.SubmissionView, error)
	ListByQuest(ctx context.Context, questID int64) ([]creative.SubmissionView, error)
	ListByCrew(ctx context.Context, crewID string) ([]creative.SubmissionView, error)
	ListByCrewAndKind(ctx context.Context, crewID, kind string) ([]creative.SubmissionView, error)
	GetSubmission(ctx context.Context, submissionID int64) (*creative.SubmissionView, error)
	Approve(ctx context.Context, submissionID int64, reviewerUID string) (*creative.SubmissionView, error)
	Reject(ctx context.Context, submissionID int64, reviewerUID string, reason string) (*creative.SubmissionView, error)
}

var handler CreativeHandler

func Setup(h CreativeHandler) {
	handler = h
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	route := parseCreativePath(r.URL.Path)

	var routeFn func(claims *auth.SessionClaims)
	switch {
	case r.Method == http.MethodPost && route.action == "" && route.idStr == "":
		routeFn = func(c *auth.SessionClaims) { handleSubmit(w, r, c) }
	case r.Method == http.MethodGet && route.action == "" && route.idStr != "":
		routeFn = func(c *auth.SessionClaims) { handleGet(w, r, c, route.idStr) }
	case r.Method == http.MethodGet && route.action == "" && route.idStr == "":
		routeFn = func(c *auth.SessionClaims) { handleList(w, r, c) }
	case r.Method == http.MethodPatch && route.action == "approve":
		routeFn = func(c *auth.SessionClaims) { handleApprove(w, r, c, route.idStr) }
	case r.Method == http.MethodPatch && route.action == "reject":
		routeFn = func(c *auth.SessionClaims) { handleReject(w, r, c, route.idStr) }
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

type routeInfo struct {
	idStr  string
	action string
}

func parseCreativePath(path string) routeInfo {
	trimmed := strings.Trim(path, "/")
	if !strings.HasPrefix(trimmed, "api/creative") {
		return routeInfo{}
	}
	rest := strings.TrimPrefix(trimmed, "api/creative")
	rest = strings.Trim(rest, "/")
	if rest == "" {
		return routeInfo{}
	}

	parts := strings.Split(rest, "/")
	if len(parts) == 1 {
		return routeInfo{idStr: parts[0]}
	}
	if len(parts) == 2 {
		return routeInfo{idStr: parts[0], action: parts[1]}
	}
	return routeInfo{}
}

func handleSubmit(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	var req creative.SubmitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sub := &game.Submission{
		MissionID:     req.MissionID,
		ExerciseID: req.ExerciseID,
		Kind:        req.Kind,
		Content:     req.Content,
	}

	view, err := handler.Submit(r.Context(), claims.UID, sub)
	if err != nil {
		switch {
		case err == creative.ErrQuestNotFound, err == creative.ErrChallengeNotFound:
			shared.WriteJSONError(w, err.Error(), http.StatusNotFound)
		case err == creative.ErrQuestNotActive:
			shared.WriteJSONError(w, err.Error(), http.StatusConflict)
		case err == creative.ErrChallengeDone:
			shared.WriteJSONError(w, err.Error(), http.StatusConflict)
		case err == creative.ErrInvalidKind, err == creative.ErrContentTooShort,
			err == creative.ErrSVGEmpty, err == creative.ErrSVGTooLarge,
			err == creative.ErrSVGMalformed, err == creative.ErrSVGRootMissing,
			err == creative.ErrSVGMultipleRoots, err == creative.ErrSVGDisallowedTag,
			err == creative.ErrSVGDisallowedAttr, err == creative.ErrSVGDisallowedURI,
			err == creative.ErrComicEmpty, err == creative.ErrComicTooLarge,
			err == creative.ErrComicMalformed, err == creative.ErrComicPanelCount,
			err == creative.ErrComicPanelEmpty, err == creative.ErrComicCaptionTooLong,
			err == creative.ErrPhotoEmpty, err == creative.ErrPhotoTooLarge,
			err == creative.ErrPhotoMalformed, err == creative.ErrPhotoMissing,
			err == creative.ErrPhotoBadURI, err == creative.ErrPhotoNotImage,
			err == creative.ErrPhotoDecode, err == creative.ErrPhotoTooBig,
			err == creative.ErrPhotoCaptionLong,
			err == creative.ErrVideoEmpty, err == creative.ErrVideoTooLarge,
			err == creative.ErrVideoMalformed, err == creative.ErrVideoMissing,
			err == creative.ErrVideoBadURI, err == creative.ErrVideoNotVideo,
			err == creative.ErrVideoDecode, err == creative.ErrVideoTooBig,
			err == creative.ErrVideoCaptionLong, err == creative.ErrVideoBadMagic:
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		default:
			shared.WriteJSONError(w, "failed to submit creative", http.StatusInternalServerError)
		}
		return
	}

	shared.WriteJSON(w, http.StatusCreated, view)
}

func handleList(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims) {
	questIDStr := r.URL.Query().Get("mission_id")

	if questIDStr != "" {
		if !isReviewer(claims) {
			shared.WriteForbidden(w)
			return
		}
		questID, err := strconv.ParseInt(questIDStr, 10, 64)
		if err != nil {
			shared.WriteJSONError(w, "invalid mission_id", http.StatusBadRequest)
			return
		}
		views, err := handler.ListByQuest(r.Context(), questID)
		if err != nil {
			shared.WriteJSONError(w, "failed to list submissions", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, views)
		return
	}

	kind := r.URL.Query().Get("kind")
	if kind != "" {
		views, err := handler.ListByCrewAndKind(r.Context(), claims.FamilyID, kind)
		if err != nil {
			shared.WriteJSONError(w, "failed to list submissions", http.StatusInternalServerError)
			return
		}
		shared.WriteJSON(w, http.StatusOK, views)
		return
	}

	views, err := handler.ListByCrew(r.Context(), claims.FamilyID)
	if err != nil {
		shared.WriteJSONError(w, "failed to list submissions", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, views)
}

func handleGet(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr string) {
	submissionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid submission id", http.StatusBadRequest)
		return
	}

	view, err := handler.GetSubmission(r.Context(), submissionID)
	if err != nil {
		if err == creative.ErrNotFound {
			shared.WriteJSONError(w, "submission not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to load submission", http.StatusInternalServerError)
		return
	}

	if view.AuthorUID != claims.UID && !isReviewer(claims) {
		shared.WriteForbidden(w)
		return
	}

	shared.WriteJSON(w, http.StatusOK, view)
}

func handleApprove(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr string) {
	if !isReviewer(claims) {
		shared.WriteForbidden(w)
		return
	}

	submissionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid submission id", http.StatusBadRequest)
		return
	}

	view, err := handler.Approve(r.Context(), submissionID, claims.UID)
	if err != nil {
		if err == creative.ErrNotFound {
			shared.WriteJSONError(w, "submission not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to approve submission", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, view)
}

func handleReject(w http.ResponseWriter, r *http.Request, claims *auth.SessionClaims, idStr string) {
	if !isReviewer(claims) {
		shared.WriteForbidden(w)
		return
	}

	submissionID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSONError(w, "invalid submission id", http.StatusBadRequest)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Reason) > 1024 {
		shared.WriteJSONError(w, "reason too long", http.StatusBadRequest)
		return
	}

	view, err := handler.Reject(r.Context(), submissionID, claims.UID, req.Reason)
	if err != nil {
		if err == creative.ErrNotFound {
			shared.WriteJSONError(w, "submission not found", http.StatusNotFound)
			return
		}
		shared.WriteJSONError(w, "failed to reject submission", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, view)
}

func isReviewer(claims *auth.SessionClaims) bool {
	switch claims.Role {
	case string(auth.RoleGuide), string(auth.RoleBuilder), string(auth.RoleAdmin):
		return true
	}
	return false
}
