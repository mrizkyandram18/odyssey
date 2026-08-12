package daily_activities

import (
	"encoding/json"
	"net/http"
	"strconv"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/dailyactivity"
	"odyssey/pkg/observability"
	"odyssey/pkg/shared"
)

type API struct {
	svc    *dailyactivity.Service
	logger *observability.Logger
}

func NewAPI(svc *dailyactivity.Service, logger *observability.Logger) *API {
	return &API{svc: svc, logger: logger}
}

func (a *API) HandleGetToday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	uid := claims.UID

	act, err := a.svc.GetToday(ctx, uid)
	if err != nil {
		if err == dailyactivity.ErrActivityNotFound {
			shared.WriteJSON(w, http.StatusNotFound, map[string]string{"error": "no activity available"})
			return
		}
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to get today's activity"})
		return
	}

	shared.WriteJSON(w, http.StatusOK, act)
}

func (a *API) HandleComplete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}
	uid := claims.UID

	idStr := r.PathValue("id")
	if idStr == "" {
		idStr = r.URL.Query().Get("id") // Fallback
	}
	activityID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid activity id"})
		return
	}

	var req struct {
		Answer string `json:"answer"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return
	}

	res, err := a.svc.CompleteActivity(ctx, uid, activityID, req.Answer)
	if err != nil {
		if err == dailyactivity.ErrAlreadyCompleted {
			shared.WriteJSON(w, http.StatusConflict, map[string]string{"error": "already completed"})
			return
		}
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to complete activity"})
		return
	}

	shared.WriteJSON(w, http.StatusOK, res)
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" && (r.URL.Path == "/api/daily-activities/today" || r.URL.Path == "/api/daily-activities/today/") {
		a.HandleGetToday(w, r)
		return
	}
	if r.Method == "POST" && len(r.URL.Path) > 22 && r.URL.Path[:22] == "/api/daily-activities/" {
		a.HandleComplete(w, r)
		return
	}
	http.NotFound(w, r)
}

func Setup(svc *dailyactivity.Service, logger *observability.Logger) *API {
	return NewAPI(svc, logger)
}
