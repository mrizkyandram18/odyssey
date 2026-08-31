package admin_tasks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
	"odyssey/pkg/tasks"
)

type API struct {
	client db.SupabaseClient
	engine *tasks.Engine
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{client: client, engine: tasks.DefaultEngine}
}

type TaskInput = tasks.TaskInput

type PendingSubmissionView struct {
	ID             int64          `json:"id"`
	TaskID         int64          `json:"task_id"`
	TaskTitle      string         `json:"task_title"`
	TaskType       string         `json:"task_type"`
	UserUID        string         `json:"user_uid"`
	UserName       string         `json:"user_name"`
	SubmissionType string         `json:"submission_type"`
	Status         string         `json:"status"`
	Payload        map[string]any `json:"payload"`
	CreatedAt      string         `json:"created_at"`
	RewardCoins    int            `json:"reward_coins"`
	RewardXP       int            `json:"reward_xp"`
}

func (a *API) requireGuide(w http.ResponseWriter, r *http.Request) (*auth.SessionClaims, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return nil, false
	}
	if claims.Role != "GUIDE" {
		shared.WriteJSONError(w, "akses ditolak: hanya untuk admin/guide", http.StatusForbidden)
		return nil, false
	}
	return claims, true
}

func (a *API) validateTaskInput(req *TaskInput) error {
	return a.engine.ValidateTaskInput(req)
}

func (a *API) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	date := r.URL.Query().Get("date")

	params := "order=active_date.desc,step_order.asc"
	if claims.FamilyID != "" {
		if date != "" {
			params = fmt.Sprintf("family_id=eq.%s&active_date=eq.%s&order=step_order.asc", claims.FamilyID, date)
		} else {
			params = fmt.Sprintf("family_id=eq.%s&order=active_date.desc,step_order.asc", claims.FamilyID)
		}
	} else if date != "" {
		params = fmt.Sprintf("active_date=eq.%s&order=step_order.asc", date)
	}

	raw, err := a.client.Get(ctx, "odyssey_tasks", params)
	if err != nil {
		shared.WriteJSONError(w, "failed to get tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Defense in depth: filter by family_id in memory if claims.FamilyID is set
	if claims.FamilyID != "" {
		var list []map[string]any
		if err := json.Unmarshal(raw, &list); err == nil {
			filtered := make([]map[string]any, 0, len(list))
			for _, item := range list {
				if famID, ok := item["family_id"].(string); ok && famID == claims.FamilyID {
					filtered = append(filtered, item)
				}
			}
			raw, _ = json.Marshal(filtered)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (a *API) HandleCreateTask(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req TaskInput
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid request: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := a.validateTaskInput(&req); err != nil {
		shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	evalType := req.EvaluationType
	if evalType == "" {
		switch req.TaskType {
		case "PHOTO_UPLOAD", "DOCUMENT_UPLOAD", "TEXT_RESPONSE", "PHOTO_PROOF":
			evalType = "ADMIN_REVIEW"
		default:
			evalType = "AUTO"
		}
	}

	payload := map[string]any{
		"title":           req.Title,
		"description":     req.Description,
		"task_type":       req.TaskType,
		"evaluation_type": evalType,
		"step_order":      req.StepOrder,
		"reward_coins":    req.RewardCoins,
		"reward_xp":       req.RewardXP,
		"config":          req.Config,
		"is_active":       req.IsActive,
		"created_by":      claims.UID,
		"family_id":       claims.FamilyID,
	}
	if req.ActiveDate != "" {
		payload["active_date"] = req.ActiveDate
	}

	raw, err := a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_tasks", payload, "", "return=representation")
	if err != nil {
		shared.WriteJSONError(w, "gagal membuat tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(raw)
}

func (a *API) HandleUpdateTask(w http.ResponseWriter, r *http.Request, taskID int64) {
	claims, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	// Verify family ownership
	tRaw, err := a.client.Get(ctx, "odyssey_tasks", fmt.Sprintf("id=eq.%d", taskID))
	if err != nil || len(tRaw) == 0 {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	type taskFamilyCheck struct {
		FamilyID string `json:"family_id"`
	}
	var checks []taskFamilyCheck
	_ = json.Unmarshal(tRaw, &checks)
	if len(checks) > 0 && claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: tugas bukan milik keluarga Anda", http.StatusForbidden)
		return
	}

	var patch map[string]any
	if err := shared.ReadJSON(r, &patch); err != nil {
		shared.WriteJSONError(w, "invalid json patch", http.StatusBadRequest)
		return
	}

	// Validate patch does not introduce invalid capability configuration
	if cfgRaw, hasCfg := patch["config"]; hasCfg && cfgRaw != nil {
		if cfgMap, ok := cfgRaw.(map[string]any); ok {
			// Determine effective task_type for validation
			effectiveType := ""
			if tt, ok := patch["task_type"].(string); ok && tt != "" {
				effectiveType = tt
			} else {
				// Fetch current task_type from DB record
				var existing []map[string]any
				_ = json.Unmarshal(tRaw, &existing)
				if len(existing) > 0 {
					if tt, ok := existing[0]["task_type"].(string); ok {
						effectiveType = tt
					}
				}
			}
			if effectiveType != "" {
				tmpInput := &tasks.TaskInput{
					Title:    "patch-validation",
					TaskType: effectiveType,
					Config:   cfgMap,
				}
				// Copy reward fields if present in patch to avoid false defaults
				if rc, ok := patch["reward_coins"].(float64); ok {
					tmpInput.RewardCoins = int(rc)
				} else {
					tmpInput.RewardCoins = 50
				}
				if rx, ok := patch["reward_xp"].(float64); ok {
					tmpInput.RewardXP = int(rx)
				} else {
					tmpInput.RewardXP = 100
				}
				tmpInput.StepOrder = 1
				if err := a.validateTaskInput(tmpInput); err != nil {
					shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
					return
				}
			}
		}
	}

	raw, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_tasks", patch, fmt.Sprintf("id=eq.%d", taskID))
	if err != nil {
		shared.WriteJSONError(w, "gagal update tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (a *API) HandleDeleteTask(w http.ResponseWriter, r *http.Request, taskID int64) {
	claims, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	// Verify family ownership
	tRaw, err := a.client.Get(ctx, "odyssey_tasks", fmt.Sprintf("id=eq.%d", taskID))
	if err != nil || len(tRaw) == 0 {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	type taskFamilyCheck struct {
		FamilyID string `json:"family_id"`
	}
	var checks []taskFamilyCheck
	_ = json.Unmarshal(tRaw, &checks)
	if len(checks) > 0 && claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: tugas bukan milik keluarga Anda", http.StatusForbidden)
		return
	}

	_, err = a.client.Mutate(ctx, http.MethodDelete, "odyssey_tasks", nil, fmt.Sprintf("id=eq.%d", taskID))
	if err != nil {
		shared.WriteJSONError(w, "gagal menghapus tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (a *API) HandleListPendingSubmissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	// 1. Fetch profiles for admin's family
	profParams := "select=uid,explorer_name,family_id"
	if claims.FamilyID != "" {
		profParams = fmt.Sprintf("family_id=eq.%s&select=uid,explorer_name,family_id", claims.FamilyID)
	}
	profRaw, _ := a.client.Get(ctx, "odyssey_user_profiles", profParams)
	type ProfRow struct {
		UID          string `json:"uid"`
		ExplorerName string `json:"explorer_name"`
		FamilyID     string `json:"family_id"`
	}
	var profs []ProfRow
	_ = json.Unmarshal(profRaw, &profs)
	profMap := make(map[string]string)
	familyMemberUIDs := make(map[string]bool)
	for _, p := range profs {
		profMap[p.UID] = p.ExplorerName
		familyMemberUIDs[p.UID] = true
	}

	// 2. Fetch pending submissions
	subRaw, err := a.client.Get(ctx, "odyssey_task_submissions", "status=eq.PENDING&order=created_at.desc")
	if err != nil {
		shared.WriteJSONError(w, "gagal mengambil submission: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type SubRow struct {
		ID             int64          `json:"id"`
		TaskID         int64          `json:"task_id"`
		UserUID        string         `json:"user_uid"`
		SubmissionType string         `json:"submission_type"`
		Status         string         `json:"status"`
		Payload        map[string]any `json:"payload"`
		CreatedAt      string         `json:"created_at"`
	}
	var subs []SubRow
	_ = json.Unmarshal(subRaw, &subs)

	// Filter by family members
	filteredSubs := make([]SubRow, 0, len(subs))
	for _, s := range subs {
		if claims.FamilyID == "" || familyMemberUIDs[s.UserUID] {
			filteredSubs = append(filteredSubs, s)
		}
	}

	if len(filteredSubs) == 0 {
		shared.WriteJSON(w, http.StatusOK, []PendingSubmissionView{})
		return
	}

	// 3. Fetch Tasks for enrichment
	taskRaw, _ := a.client.Get(ctx, "odyssey_tasks", "select=id,title,task_type,reward_coins,reward_xp")
	type TaskRow struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		TaskType    string `json:"task_type"`
		RewardCoins int    `json:"reward_coins"`
		RewardXP    int    `json:"reward_xp"`
	}
	var tasks []TaskRow
	_ = json.Unmarshal(taskRaw, &tasks)
	taskMap := make(map[int64]TaskRow)
	for _, t := range tasks {
		taskMap[t.ID] = t
	}

	// 4. Build enriched response
	res := make([]PendingSubmissionView, len(filteredSubs))
	for i, s := range filteredSubs {
		t := taskMap[s.TaskID]
		name := profMap[s.UserUID]
		if name == "" {
			name = s.UserUID
		}
		res[i] = PendingSubmissionView{
			ID:             s.ID,
			TaskID:         s.TaskID,
			TaskTitle:      t.Title,
			TaskType:       t.TaskType,
			UserUID:        s.UserUID,
			UserName:       name,
			SubmissionType: s.SubmissionType,
			Status:         s.Status,
			Payload:        s.Payload,
			CreatedAt:      s.CreatedAt,
			RewardCoins:    t.RewardCoins,
			RewardXP:       t.RewardXP,
		}
	}

	shared.WriteJSON(w, http.StatusOK, res)
}

func (a *API) HandleVerifySubmission(w http.ResponseWriter, r *http.Request, subID int64) {
	claims, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req struct {
		Status string `json:"status"` // "APPROVED" or "REJECTED"
		Notes  string `json:"notes"`
	}
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if req.Status != "APPROVED" && req.Status != "REJECTED" {
		shared.WriteJSONError(w, "status must be APPROVED or REJECTED", http.StatusBadRequest)
		return
	}

	rpcRes, err := a.client.RPC(ctx, "odyssey_verify_submission", map[string]any{
		"p_submission_id": subID,
		"p_admin_uid":     claims.UID,
		"p_status":        req.Status,
		"p_admin_notes":   req.Notes,
	})
	if err != nil {
		shared.WriteJSONError(w, "gagal verifikasi submission: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rpcRes)
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	// Submissions queue & verify
	if r.Method == http.MethodGet && path == "/api/admin/submissions/pending" {
		a.HandleListPendingSubmissions(w, r)
		return
	}
	if r.Method == http.MethodPost && strings.HasPrefix(path, "/api/admin/submissions/") && strings.HasSuffix(path, "/verify") {
		parts := strings.Split(path, "/")
		if len(parts) >= 5 {
			subID, err := strconv.ParseInt(parts[4], 10, 64)
			if err == nil {
				a.HandleVerifySubmission(w, r, subID)
				return
			}
		}
	}

	// Redemption config management
	if path == "/api/admin/config" {
		if r.Method == http.MethodGet {
			a.HandleGetAdminConfig(w, r)
			return
		}
		if r.Method == http.MethodPost || r.Method == http.MethodPatch || r.Method == http.MethodPut {
			a.HandleUpdateAdminConfig(w, r)
			return
		}
	}

	// Task management
	if path == "/api/admin/tasks" {
		if r.Method == http.MethodGet {
			a.HandleListTasks(w, r)
			return
		}
		if r.Method == http.MethodPost {
			a.HandleCreateTask(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/api/admin/tasks/") {
		taskIDStr := strings.TrimPrefix(path, "/api/admin/tasks/")
		taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
		if err == nil {
			if r.Method == http.MethodPatch {
				a.HandleUpdateTask(w, r, taskID)
				return
			}
			if r.Method == http.MethodDelete {
				a.HandleDeleteTask(w, r, taskID)
				return
			}
		}
	}

	http.NotFound(w, r)
}

func (a *API) HandleGetAdminConfig(w http.ResponseWriter, r *http.Request) {
	_, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	startDay := shared.DefaultRedemptionStartDay
	endDay := shared.DefaultRedemptionEndDay
	raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(redemption_start_day,redemption_end_day)")
	if err == nil && len(raw) > 0 {
		type ConfigRow struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		var rows []ConfigRow
		if err := json.Unmarshal(raw, &rows); err == nil {
			for _, row := range rows {
				if row.Key == "redemption_start_day" {
					if v, err := strconv.Atoi(row.Value); err == nil && v >= 1 && v <= 31 {
						startDay = v
					}
				} else if row.Key == "redemption_end_day" {
					if v, err := strconv.Atoi(row.Value); err == nil && v >= 1 && v <= 31 {
						endDay = v
					}
				}
			}
		}
	}
	tz := shared.LoadConfig().Timezone
	cfg := shared.ResolveRedemptionConfig(startDay, endDay, tz, time.Now())
	shared.WriteJSON(w, http.StatusOK, cfg)
}

func (a *API) HandleUpdateAdminConfig(w http.ResponseWriter, r *http.Request) {
	_, ok := a.requireGuide(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req struct {
		StartDay int `json:"start_day"`
		EndDay   int `json:"end_day"`
	}
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.StartDay < 1 || req.StartDay > 31 || req.EndDay < 1 || req.EndDay > 31 || req.StartDay > req.EndDay {
		shared.WriteJSONError(w, "rentang tanggal penukaran tidak valid (1 <= start_day <= end_day <= 31)", http.StatusBadRequest)
		return
	}

	startPayload := map[string]any{
		"key":   "redemption_start_day",
		"value": strconv.Itoa(req.StartDay),
	}
	endPayload := map[string]any{
		"key":   "redemption_end_day",
		"value": strconv.Itoa(req.EndDay),
	}

	_, err := a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_system_config", startPayload, "", "resolution=merge-duplicates")
	if err != nil {
		_, _ = a.client.Mutate(ctx, http.MethodPatch, "odyssey_system_config", map[string]any{"value": strconv.Itoa(req.StartDay)}, "key=eq.redemption_start_day")
	}
	_, err = a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_system_config", endPayload, "", "resolution=merge-duplicates")
	if err != nil {
		_, _ = a.client.Mutate(ctx, http.MethodPatch, "odyssey_system_config", map[string]any{"value": strconv.Itoa(req.EndDay)}, "key=eq.redemption_end_day")
	}

	tz := shared.LoadConfig().Timezone
	cfg := shared.ResolveRedemptionConfig(req.StartDay, req.EndDay, tz, time.Now())
	shared.WriteJSON(w, http.StatusOK, cfg)
}
