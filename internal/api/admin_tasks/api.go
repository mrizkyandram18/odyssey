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
	CoinsEarned    int            `json:"coins_earned,omitempty"`
	XPEarned       int            `json:"xp_earned,omitempty"`
	AdminNotes     *string        `json:"admin_notes,omitempty"`
	ReviewedAt     *string        `json:"reviewed_at,omitempty"`
}

func (a *API) requireAdmin(w http.ResponseWriter, r *http.Request) (*auth.SessionClaims, bool) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return nil, false
	}
	role := auth.NormalizeRole(claims.Role)
	if role != auth.RoleAdmin {
		shared.WriteJSONError(w, "akses ditolak: hanya untuk admin", http.StatusForbidden)
		return nil, false
	}
	return claims, true
}

func (a *API) validateTaskInput(req *TaskInput) error {
	return a.engine.ValidateTaskInput(req)
}

func (a *API) HandleListTasks(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	date := strings.TrimSpace(r.URL.Query().Get("date"))
	// Defensive default: empty date must not silently return all tasks; default to today in system timezone
	if date == "" {
		tz := shared.LoadConfig().Timezone
		if tz == "" {
			tz = shared.DefaultTimezone
		}
		if raw, err := a.client.Get(ctx, "odyssey_system_config", "key=eq.timezone&select=value"); err == nil && len(raw) > 0 {
			var rows []struct {
				Value string `json:"value"`
			}
			if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 && strings.TrimSpace(rows[0].Value) != "" {
				if _, err := time.LoadLocation(strings.TrimSpace(rows[0].Value)); err == nil {
					tz = strings.TrimSpace(rows[0].Value)
				}
			}
		}
		loc, _ := time.LoadLocation(tz)
		if loc == nil {
			loc = time.FixedZone("WIB", 7*3600)
		}
		date = time.Now().In(loc).Format("2006-01-02")
	}

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
	claims, ok := a.requireAdmin(w, r)
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

	targetScope := strings.ToUpper(strings.TrimSpace(req.TargetScope))
	if targetScope == "" {
		targetScope = "ALL"
	}
	if targetScope != "ALL" && targetScope != "FAMILY" && targetScope != "USER" {
		shared.WriteJSONError(w, "target_scope tidak valid", http.StatusBadRequest)
		return
	}

	targetUID := strings.TrimSpace(req.TargetUserUID)
	if targetScope == "USER" {
		if targetUID == "" {
			shared.WriteJSONError(w, "target_user_uid wajib diisi untuk target user tertentu", http.StatusBadRequest)
			return
		}
		// Validate target user exists and belongs to admin's family
		uRaw, err := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
		if err != nil || len(uRaw) == 0 {
			shared.WriteJSONError(w, "user target tidak ditemukan", http.StatusBadRequest)
			return
		}
		type UserCheck struct {
			FamilyID string `json:"family_id"`
		}
		var uChecks []UserCheck
		_ = json.Unmarshal(uRaw, &uChecks)
		if len(uChecks) > 0 && claims.FamilyID != "" && uChecks[0].FamilyID != "" && uChecks[0].FamilyID != claims.FamilyID {
			shared.WriteJSONError(w, "akses ditolak: user target bukan milik keluarga Anda", http.StatusForbidden)
			return
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
		"target_scope":    targetScope,
		"is_active":       req.IsActive,
		"created_by":      claims.UID,
		"family_id":       claims.FamilyID,
	}
	if targetScope == "USER" {
		payload["target_user_uid"] = targetUID
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
	claims, ok := a.requireAdmin(w, r)
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

	// Submission guard: task_type is immutable once submissions exist
	if newTypeRaw, hasType := patch["task_type"]; hasType {
		if newTypeStr, ok := newTypeRaw.(string); ok && newTypeStr != "" {
			var existing []map[string]any
			_ = json.Unmarshal(tRaw, &existing)
			existingType := ""
			if len(existing) > 0 {
				if tt, ok := existing[0]["task_type"].(string); ok {
					existingType = tt
				}
			}
			if existingType != "" && newTypeStr != existingType {
				subRaw, _ := a.client.Get(ctx, "odyssey_task_submissions", fmt.Sprintf("task_id=eq.%d&select=id&limit=1", taskID))
				var subs []map[string]any
				_ = json.Unmarshal(subRaw, &subs)
				if len(subs) > 0 {
					shared.WriteJSONError(w, "Jenis tugas tidak dapat diubah karena tugas ini sudah memiliki submission. Buat tugas baru jika ingin mengganti jenis tugas.", http.StatusBadRequest)
					return
				}
			}
		}
	}

	// Validate effective task_type + config atomically (always, not only when config present)
	effectiveType := ""
	effectiveConfig := map[string]any{}
	// Resolve effective type
	if tt, ok := patch["task_type"].(string); ok && tt != "" {
		effectiveType = tt
	} else {
		var existing []map[string]any
		_ = json.Unmarshal(tRaw, &existing)
		if len(existing) > 0 {
			if tt, ok := existing[0]["task_type"].(string); ok {
				effectiveType = tt
			}
		}
	}
	// Resolve effective config
	if cfgRaw, hasCfg := patch["config"]; hasCfg {
		if cfgMap, ok := cfgRaw.(map[string]any); ok {
			effectiveConfig = cfgMap
		} else if cfgRaw == nil {
			effectiveConfig = map[string]any{}
		}
	} else {
		var existing []map[string]any
		_ = json.Unmarshal(tRaw, &existing)
		if len(existing) > 0 {
			if cfg, ok := existing[0]["config"].(map[string]any); ok {
				effectiveConfig = cfg
			}
		}
	}
	if effectiveType != "" {
		tmpInput := &tasks.TaskInput{
			Title:    "patch-validation",
			TaskType: effectiveType,
			Config:   effectiveConfig,
		}
		if rc, ok := patch["reward_coins"].(float64); ok {
			tmpInput.RewardCoins = int(rc)
		} else {
			// Use existing reward for validation fallback
			var existing []map[string]any
			_ = json.Unmarshal(tRaw, &existing)
			if len(existing) > 0 {
				if rc2, ok := existing[0]["reward_coins"].(float64); ok {
					tmpInput.RewardCoins = int(rc2)
				} else {
					tmpInput.RewardCoins = 50
				}
			} else {
				tmpInput.RewardCoins = 50
			}
		}
		if rx, ok := patch["reward_xp"].(float64); ok {
			tmpInput.RewardXP = int(rx)
		} else {
			var existing []map[string]any
			_ = json.Unmarshal(tRaw, &existing)
			if len(existing) > 0 {
				if rx2, ok := existing[0]["reward_xp"].(float64); ok {
					tmpInput.RewardXP = int(rx2)
				} else {
					tmpInput.RewardXP = 100
				}
			} else {
				tmpInput.RewardXP = 100
			}
		}
		tmpInput.StepOrder = 1
		if err := a.validateTaskInput(tmpInput); err != nil {
			shared.WriteJSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	raw, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_tasks", patch, fmt.Sprintf("id=eq.%d", taskID))
	if err != nil {
		shared.WriteJSONError(w, "gagal update tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Supabase PATCH returns empty body without Prefer: return=representation
	if len(raw) == 0 || string(raw) == "[]" || string(raw) == "null" {
		shared.WriteJSON(w, http.StatusOK, map[string]any{"status": "updated", "id": taskID})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (a *API) HandleDeleteTask(w http.ResponseWriter, r *http.Request, taskID int64) {
	claims, ok := a.requireAdmin(w, r)
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

func (a *API) HandleDuplicateTask(w http.ResponseWriter, r *http.Request, taskID int64) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	tRaw, err := a.client.Get(ctx, "odyssey_tasks", fmt.Sprintf("id=eq.%d", taskID))
	if err != nil || len(tRaw) == 0 {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	var list []map[string]any
	_ = json.Unmarshal(tRaw, &list)
	if len(list) == 0 {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	orig := list[0]

	if claims.FamilyID != "" {
		if famID, ok := orig["family_id"].(string); ok && famID != "" && famID != claims.FamilyID {
			shared.WriteJSONError(w, "akses ditolak: tugas bukan milik keluarga Anda", http.StatusForbidden)
			return
		}
	}

	newTitle := fmt.Sprintf("%v (Salinan)", orig["title"])
	payload := map[string]any{
		"title":           newTitle,
		"description":     orig["description"],
		"task_type":       orig["task_type"],
		"evaluation_type": orig["evaluation_type"],
		"step_order":      orig["step_order"],
		"reward_coins":    orig["reward_coins"],
		"reward_xp":       orig["reward_xp"],
		"config":          orig["config"],
		"is_active":       orig["is_active"],
		"created_by":      claims.UID,
		"family_id":       claims.FamilyID,
	}
	if actDate, ok := orig["active_date"].(string); ok && actDate != "" {
		payload["active_date"] = actDate
	}
	if scope, ok := orig["target_scope"].(string); ok && scope != "" {
		payload["target_scope"] = scope
	}
	if targetUID, ok := orig["target_user_uid"].(string); ok && targetUID != "" {
		payload["target_user_uid"] = targetUID
	}

	raw, err := a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_tasks", payload, "", "return=representation")
	if err != nil {
		shared.WriteJSONError(w, "gagal duplikasi tugas: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(raw)
}

func (a *API) HandleListPendingSubmissions(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	page, limit, offset := shared.ParsePagination(r, 50, 100)

	familyID := claims.FamilyID
	if familyID == "" {
		familyID = "family_default"
	}

	// 1. Fetch profiles strictly scoped to admin's family
	profParams := fmt.Sprintf("family_id=eq.%s&select=uid,explorer_name", familyID)
	profRaw, _ := a.client.Get(ctx, "odyssey_user_profiles", profParams)
	type ProfRow struct {
		UID          string `json:"uid"`
		ExplorerName string `json:"explorer_name"`
	}
	var profs []ProfRow
	_ = json.Unmarshal(profRaw, &profs)

	if len(profs) == 0 {
		shared.WriteJSON(w, http.StatusOK, shared.PaginatedResponse[PendingSubmissionView]{
			Items: []PendingSubmissionView{},
			Pagination: shared.PaginationMeta{
				Page:    page,
				Limit:   limit,
				Total:   0,
				HasNext: false,
			},
		})
		return
	}

	profMap := make(map[string]string, len(profs))
	uids := make([]string, len(profs))
	for i, p := range profs {
		profMap[p.UID] = p.ExplorerName
		uids[i] = p.UID
	}

	// 2. Determine status filter
	statusParam := strings.ToUpper(strings.TrimSpace(r.URL.Query().Get("status")))
	statusClause := ""
	if statusParam == "PENDING" || statusParam == "APPROVED" || statusParam == "REJECTED" {
		statusClause = fmt.Sprintf("&status=eq.%s", statusParam)
	} else if statusParam == "" && strings.HasSuffix(r.URL.Path, "/pending") {
		statusClause = "&status=eq.PENDING"
	}

	// 3. Fetch submissions strictly scoped to family member UIDs with deterministic pagination
	subParams := fmt.Sprintf("user_uid=in.(%s)%s&order=created_at.desc,id.desc&limit=%d&offset=%d", strings.Join(uids, ","), statusClause, limit, offset)

	subRaw, err := a.client.Get(ctx, "odyssey_task_submissions", subParams)
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
		CoinsEarned    int            `json:"coins_earned"`
		XPEarned       int            `json:"xp_earned"`
		AdminNotes     *string        `json:"admin_notes"`
		ReviewedAt     *string        `json:"reviewed_at"`
	}
	var subs []SubRow
	_ = json.Unmarshal(subRaw, &subs)

	if len(subs) == 0 {
		shared.WriteJSON(w, http.StatusOK, shared.PaginatedResponse[PendingSubmissionView]{
			Items: []PendingSubmissionView{},
			Pagination: shared.PaginationMeta{
				Page:    page,
				Limit:   limit,
				Total:   0,
				HasNext: false,
			},
		})
		return
	}

	// 4. Fetch Tasks ONLY for task IDs in the current paginated page (targeted id=in.(...))
	taskIDSet := make(map[int64]bool)
	var taskIDStrs []string
	for _, s := range subs {
		if !taskIDSet[s.TaskID] {
			taskIDSet[s.TaskID] = true
			taskIDStrs = append(taskIDStrs, strconv.FormatInt(s.TaskID, 10))
		}
	}

	type TaskRow struct {
		ID          int64  `json:"id"`
		Title       string `json:"title"`
		TaskType    string `json:"task_type"`
		RewardCoins int    `json:"reward_coins"`
		RewardXP    int    `json:"reward_xp"`
	}
	taskMap := make(map[int64]TaskRow)
	if len(taskIDStrs) > 0 {
		taskParams := fmt.Sprintf("id=in.(%s)&select=id,title,task_type,reward_coins,reward_xp", strings.Join(taskIDStrs, ","))
		taskRaw, _ := a.client.Get(ctx, "odyssey_tasks", taskParams)
		var tasksList []TaskRow
		_ = json.Unmarshal(taskRaw, &tasksList)
		for _, t := range tasksList {
			taskMap[t.ID] = t
		}
	}

	// 5. Build enriched response
	items := make([]PendingSubmissionView, len(subs))
	for i, s := range subs {
		t := taskMap[s.TaskID]
		name := profMap[s.UserUID]
		if name == "" {
			name = s.UserUID
		}
		rewardCoins := t.RewardCoins
		if s.CoinsEarned > 0 {
			rewardCoins = s.CoinsEarned
		}
		rewardXP := t.RewardXP
		if s.XPEarned > 0 {
			rewardXP = s.XPEarned
		}

		items[i] = PendingSubmissionView{
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
			RewardCoins:    rewardCoins,
			RewardXP:       rewardXP,
			CoinsEarned:    s.CoinsEarned,
			XPEarned:       s.XPEarned,
			AdminNotes:     s.AdminNotes,
			ReviewedAt:     s.ReviewedAt,
		}
	}

	hasNext := len(subs) == limit
	total := offset + len(items)
	if hasNext {
		total += 1
	}

	shared.WriteJSON(w, http.StatusOK, shared.PaginatedResponse[PendingSubmissionView]{
		Items: items,
		Pagination: shared.PaginationMeta{
			Page:    page,
			Limit:   limit,
			Total:   total,
			HasNext: hasNext,
		},
	})
}

func (a *API) HandleVerifySubmission(w http.ResponseWriter, r *http.Request, subID int64) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req struct {
		Status       string `json:"status"` // "APPROVED" or "REJECTED"
		Notes        string `json:"notes"`
		PenaltyCoins int    `json:"penalty_coins,omitempty"`
	}
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if req.Status != "APPROVED" && req.Status != "REJECTED" {
		shared.WriteJSONError(w, "status must be APPROVED or REJECTED", http.StatusBadRequest)
		return
	}
	if req.PenaltyCoins < 0 {
		shared.WriteJSONError(w, "penalty_coins must be non-negative", http.StatusBadRequest)
		return
	}
	if req.Status == "APPROVED" && req.PenaltyCoins > 0 {
		shared.WriteJSONError(w, "penalty_coins cannot be applied to approved submissions", http.StatusBadRequest)
		return
	}

	rpcRes, err := a.client.RPC(ctx, "odyssey_verify_submission", map[string]any{
		"p_submission_id": subID,
		"p_admin_uid":     claims.UID,
		"p_status":        req.Status,
		"p_admin_notes":   req.Notes,
		"p_penalty_coins": req.PenaltyCoins,
	})
	if err != nil {
		shared.WriteJSONError(w, "gagal verifikasi submission: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rpcRes)
}

func (a *API) HandleEditSubmission(w http.ResponseWriter, r *http.Request, subID int64) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req struct {
		Payload map[string]any `json:"payload"`
		Notes   *string        `json:"notes,omitempty"`
	}
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json payload", http.StatusBadRequest)
		return
	}

	if req.Payload == nil {
		shared.WriteJSONError(w, "payload is required", http.StatusBadRequest)
		return
	}

	// Sanitize / validate known payload fields
	if textVal, ok := req.Payload["text"].(string); ok {
		if len(textVal) > 10000 {
			shared.WriteJSONError(w, "text payload exceeds maximum character limit (10000)", http.StatusBadRequest)
			return
		}
	}
	if scoreVal, ok := req.Payload["score"].(float64); ok {
		if scoreVal < 0 || scoreVal > 1000000 {
			shared.WriteJSONError(w, "score must be between 0 and 1,000,000", http.StatusBadRequest)
			return
		}
	}

	var notesParam any
	if req.Notes != nil {
		notesParam = *req.Notes
	}

	rpcRes, err := a.client.RPC(ctx, "odyssey_admin_edit_submission", map[string]any{
		"p_submission_id": subID,
		"p_admin_uid":     claims.UID,
		"p_payload":       req.Payload,
		"p_admin_notes":   notesParam,
	})
	if err != nil {
		shared.WriteJSONError(w, "gagal mengedit submission: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rpcRes)
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	// Submissions queue & verify & edit
	if r.Method == http.MethodGet && (path == "/api/admin/submissions/pending" || path == "/api/admin/submissions") {
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
	if r.Method == http.MethodPatch && strings.HasPrefix(path, "/api/admin/submissions/") {
		parts := strings.Split(path, "/")
		if len(parts) == 5 {
			subID, err := strconv.ParseInt(parts[4], 10, 64)
			if err == nil {
				a.HandleEditSubmission(w, r, subID)
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
		if strings.HasSuffix(path, "/duplicate") && r.Method == http.MethodPost {
			subPath := strings.TrimPrefix(path, "/api/admin/tasks/")
			taskIDStr := strings.TrimSuffix(subPath, "/duplicate")
			if taskID, err := strconv.ParseInt(taskIDStr, 10, 64); err == nil {
				a.HandleDuplicateTask(w, r, taskID)
				return
			}
		}
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
	_, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	startDay := shared.DefaultRedemptionStartDay
	endDay := shared.DefaultRedemptionEndDay
	payoutDay := shared.DefaultPayoutDay
	earningPeriodDays := shared.DefaultEarningPeriodDays
	conversionRate := shared.DefaultCoinConversionRate
	payoutTargetRupiah := shared.DefaultPayoutTargetRupiah
	payoutTargetCoins := shared.DefaultPayoutTargetCoins
	maxPayoutCoins := shared.DefaultMaxPayoutCoins
	timezone := shared.DefaultTimezone
	autoBlockDays := shared.DefaultAutoBlockInactivityDays
	raw, err := a.client.Get(ctx, "odyssey_system_config", "key=in.(redemption_start_day,redemption_end_day,payout_day,earning_period_days,coin_conversion_rate,payout_target_rupiah,payout_target_coins,max_payout_coins,timezone,auto_block_inactivity_days,AUTO_BLOCK_INACTIVITY_DAYS)")
	if err == nil && len(raw) > 0 {
		type ConfigRow struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		var rows []ConfigRow
		if err := json.Unmarshal(raw, &rows); err == nil {
			for _, row := range rows {
				switch row.Key {
				case "redemption_start_day":
					if v, err := strconv.Atoi(row.Value); err == nil && v >= 1 && v <= 31 {
						startDay = v
					}
				case "redemption_end_day":
					if v, err := strconv.Atoi(row.Value); err == nil && v >= 1 && v <= 31 {
						endDay = v
					}
				case "payout_day":
					if v, err := strconv.Atoi(row.Value); err == nil && v >= 1 && v <= 31 {
						payoutDay = v
					}
				case "earning_period_days":
					if v, err := strconv.Atoi(row.Value); err == nil && v > 0 {
						earningPeriodDays = v
					}
				case "coin_conversion_rate":
					if v, err := strconv.Atoi(row.Value); err == nil && v > 0 {
						conversionRate = v
					}
				case "payout_target_rupiah":
					if v, err := strconv.Atoi(row.Value); err == nil && v >= 0 {
						payoutTargetRupiah = v
					}
				case "payout_target_coins":
					if v, err := strconv.Atoi(row.Value); err == nil && v > 0 {
						payoutTargetCoins = v
					}
				case "max_payout_coins":
					if v, err := strconv.Atoi(row.Value); err == nil && v > 0 {
						maxPayoutCoins = v
					}
				case "timezone":
					if v := strings.TrimSpace(row.Value); v != "" {
						if _, err := time.LoadLocation(v); err == nil {
							timezone = v
						}
					}
				case "auto_block_inactivity_days", "AUTO_BLOCK_INACTIVITY_DAYS":
					if v, err := strconv.Atoi(strings.TrimSpace(row.Value)); err == nil && v >= 0 && v <= 365 {
						autoBlockDays = v
						if v == 0 {
							autoBlockDays = 0
						}
					}
				}
			}
		}
	}
	if timezone == shared.DefaultTimezone {
		if envTZ := shared.LoadConfig().Timezone; envTZ != "" {
			timezone = envTZ
		}
	}
	cfg := shared.ResolveRedemptionConfigFull(shared.ResolveRedemptionConfigParams{
		StartDay:                startDay,
		EndDay:                  endDay,
		PayoutDay:               payoutDay,
		EarningPeriodDays:       earningPeriodDays,
		Timezone:                timezone,
		Now:                     time.Now(),
		ConversionRate:          conversionRate,
		PayoutTargetRupiah:      payoutTargetRupiah,
		PayoutTargetCoins:       payoutTargetCoins,
		MaxPayoutCoins:          maxPayoutCoins,
		AutoBlockInactivityDays: autoBlockDays,
	})
	shared.WriteJSON(w, http.StatusOK, cfg)
}

func (a *API) HandleUpdateAdminConfig(w http.ResponseWriter, r *http.Request) {
	_, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req struct {
		StartDay                *int    `json:"start_day"`
		EndDay                  *int    `json:"end_day"`
		PayoutDay               *int    `json:"payout_day"`
		EarningPeriodDays       *int    `json:"earning_period_days"`
		ConversionRate          *int    `json:"conversion_rate"`
		PayoutTargetRupiah      *int    `json:"payout_target_rupiah"`
		PayoutTargetCoins       *int    `json:"payout_target_coins"`
		MaxPayoutCoins          *int    `json:"max_payout_coins"`
		Timezone                *string `json:"timezone"`
		AutoBlockInactivityDays *int    `json:"auto_block_inactivity_days"`
	}
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid request payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Fetch current config to merge
	currentRaw, _ := a.client.Get(ctx, "odyssey_system_config", "key=in.(redemption_start_day,redemption_end_day,payout_day,earning_period_days,coin_conversion_rate,payout_target_rupiah,max_payout_coins,timezone)")
	curMap := map[string]string{}
	if len(currentRaw) > 0 {
		var typed []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		}
		if err := json.Unmarshal(currentRaw, &typed); err == nil {
			for _, r := range typed {
				curMap[r.Key] = r.Value
			}
		}
	}

	// Determine effective range for validation
	effectiveStart := shared.DefaultRedemptionStartDay
	effectiveEnd := shared.DefaultRedemptionEndDay
	if v, err := strconv.Atoi(curMap["redemption_start_day"]); err == nil {
		effectiveStart = v
	}
	if v, err := strconv.Atoi(curMap["redemption_end_day"]); err == nil {
		effectiveEnd = v
	}
	if req.StartDay != nil {
		effectiveStart = *req.StartDay
	}
	if req.EndDay != nil {
		effectiveEnd = *req.EndDay
	}
	if effectiveStart > effectiveEnd {
		shared.WriteJSONError(w, "start_day cannot be greater than end_day", http.StatusBadRequest)
		return
	}

	// Validate target coins <= max payout
	effectiveRate := shared.DefaultCoinConversionRate
	if v, err := strconv.Atoi(curMap["coin_conversion_rate"]); err == nil {
		effectiveRate = v
	}
	if req.ConversionRate != nil {
		effectiveRate = *req.ConversionRate
	}

	effectiveTargetRupiah := shared.DefaultPayoutTargetRupiah
	if v, err := strconv.Atoi(curMap["payout_target_rupiah"]); err == nil {
		effectiveTargetRupiah = v
	}
	if req.PayoutTargetRupiah != nil {
		effectiveTargetRupiah = *req.PayoutTargetRupiah
	}

	effectiveMaxPayout := shared.DefaultMaxPayoutCoins
	if v, err := strconv.Atoi(curMap["max_payout_coins"]); err == nil {
		effectiveMaxPayout = v
	}
	if req.MaxPayoutCoins != nil {
		effectiveMaxPayout = *req.MaxPayoutCoins
	}

	if effectiveRate > 0 && effectiveTargetRupiah >= 0 {
		effectiveTargetCoins := effectiveTargetRupiah / effectiveRate
		if effectiveTargetCoins > effectiveMaxPayout {
			shared.WriteJSONError(w, fmt.Sprintf("target payout (%d coins) cannot exceed max payout (%d coins)", effectiveTargetCoins, effectiveMaxPayout), http.StatusBadRequest)
			return
		}
	}

	// Helper to upsert key
	upsert := func(key, value string) {
		payload := map[string]any{"key": key, "value": value}
		_, err := a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_system_config", payload, "", "resolution=merge-duplicates")
		if err != nil {
			_, _ = a.client.Mutate(ctx, http.MethodPatch, "odyssey_system_config", map[string]any{"value": value}, "key=eq."+key)
		}
	}

	if req.StartDay != nil {
		if *req.StartDay < 1 || *req.StartDay > 31 {
			shared.WriteJSONError(w, "start_day must be 1-31", http.StatusBadRequest)
			return
		}
		upsert("redemption_start_day", strconv.Itoa(*req.StartDay))
	}
	if req.EndDay != nil {
		if *req.EndDay < 1 || *req.EndDay > 31 {
			shared.WriteJSONError(w, "end_day must be 1-31", http.StatusBadRequest)
			return
		}
		upsert("redemption_end_day", strconv.Itoa(*req.EndDay))
	}
	if req.PayoutDay != nil {
		if *req.PayoutDay < 1 || *req.PayoutDay > 31 {
			shared.WriteJSONError(w, "payout_day must be 1-31", http.StatusBadRequest)
			return
		}
		upsert("payout_day", strconv.Itoa(*req.PayoutDay))
	}
	if req.EarningPeriodDays != nil {
		if *req.EarningPeriodDays < 1 || *req.EarningPeriodDays > 365 {
			shared.WriteJSONError(w, "earning_period_days must be 1-365", http.StatusBadRequest)
			return
		}
		upsert("earning_period_days", strconv.Itoa(*req.EarningPeriodDays))
	}
	if req.ConversionRate != nil {
		if *req.ConversionRate <= 0 {
			shared.WriteJSONError(w, "conversion_rate must be > 0", http.StatusBadRequest)
			return
		}
		upsert("coin_conversion_rate", strconv.Itoa(*req.ConversionRate))
	}
	if req.PayoutTargetRupiah != nil {
		if *req.PayoutTargetRupiah < 0 {
			shared.WriteJSONError(w, "payout_target_rupiah must be >= 0", http.StatusBadRequest)
			return
		}
		upsert("payout_target_rupiah", strconv.Itoa(*req.PayoutTargetRupiah))
	}
	if req.MaxPayoutCoins != nil {
		if *req.MaxPayoutCoins <= 0 {
			shared.WriteJSONError(w, "max_payout_coins must be > 0", http.StatusBadRequest)
			return
		}
		upsert("max_payout_coins", strconv.Itoa(*req.MaxPayoutCoins))
	}
	if req.Timezone != nil {
		if _, err := time.LoadLocation(*req.Timezone); err != nil {
			shared.WriteJSONError(w, "invalid timezone", http.StatusBadRequest)
			return
		}
		upsert("timezone", *req.Timezone)
	}
	if req.AutoBlockInactivityDays != nil {
		if *req.AutoBlockInactivityDays < 0 || *req.AutoBlockInactivityDays > 365 {
			shared.WriteJSONError(w, "auto_block_inactivity_days must be 0..365 (0 = disabled)", http.StatusBadRequest)
			return
		}
		upsert("auto_block_inactivity_days", strconv.Itoa(*req.AutoBlockInactivityDays))
		upsert("AUTO_BLOCK_INACTIVITY_DAYS", strconv.Itoa(*req.AutoBlockInactivityDays))
	}

	// Return updated config
	a.HandleGetAdminConfig(w, r)
}
