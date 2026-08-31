package family_tasks

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
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
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{client: client}
}

type TaskRecord struct {
	ID             int64          `json:"id"`
	FamilyID       string         `json:"family_id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	TaskType       string         `json:"task_type"`
	EvaluationType string         `json:"evaluation_type,omitempty"`
	StepOrder      int            `json:"step_order"`
	RewardCoins    int            `json:"reward_coins"`
	RewardXP       int            `json:"reward_xp"`
	YoutubeURL     string         `json:"youtube_url,omitempty"`
	Questions      any            `json:"questions,omitempty"`
	Config         map[string]any `json:"config,omitempty"`
	IsActive       bool           `json:"is_active"`
}

type SubmissionRecord struct {
	ID             int64   `json:"id"`
	TaskID         int64   `json:"task_id"`
	UserUID        string  `json:"user_uid"`
	SubmissionType string  `json:"submission_type"`
	Status         string  `json:"status"`
	CoinsEarned    int     `json:"coins_earned"`
	XPEarned       int     `json:"xp_earned"`
	AdminNotes     *string `json:"admin_notes,omitempty"`
	CreatedAt      string  `json:"created_at"`
}

type TaskView struct {
	ID             int64          `json:"id"`
	Title          string         `json:"title"`
	Description    string         `json:"description"`
	TaskType       string         `json:"task_type"`
	EvaluationType string         `json:"evaluation_type,omitempty"`
	StepOrder      int            `json:"step_order"`
	RewardCoins    int            `json:"reward_coins"`
	RewardXP       int            `json:"reward_xp"`
	Status         string         `json:"status"` // "UNLOCKED", "LOCKED", "PENDING", "APPROVED", "REJECTED"
	IsLocked       bool           `json:"is_locked"`
	Config         map[string]any `json:"config,omitempty"`
	AdminNotes     *string        `json:"admin_notes,omitempty"`
	CoinsEarned    int            `json:"coins_earned"`
	XPEarned       int            `json:"xp_earned"`
	SubmittedAt    *string        `json:"submitted_at,omitempty"`
}

// sanitizeValue recursively strips any answer-leaking fields from maps, slices, and nested structures.
func sanitizeValue(val any) any {
	if val == nil {
		return nil
	}
	switch v := val.(type) {
	case map[string]any:
		cleaned := make(map[string]any, len(v))
		for k, item := range v {
			lowerK := strings.ToLower(k)
			// Normalize: remove underscores for camelCase / snake_case agnostic matching
			normalized := strings.ReplaceAll(lowerK, "_", "")
			if lowerK == "correct_answer" || normalized == "correctanswer" ||
				lowerK == "correct_ans" || normalized == "correctans" ||
				lowerK == "expected_answer" || normalized == "expectedanswer" ||
				lowerK == "answer_key" || normalized == "answerkey" ||
				lowerK == "is_correct" || normalized == "iscorrect" ||
				lowerK == "answer" || lowerK == "solution" || lowerK == "correct" ||
				lowerK == "correct_option" || normalized == "correctoption" ||
				strings.Contains(normalized, "correctanswer") ||
				strings.Contains(normalized, "answerkey") ||
				strings.Contains(lowerK, "is_correct") {
				continue
			}
			cleaned[k] = sanitizeValue(item)
		}
		return cleaned
	case []any:
		cleanedList := make([]any, len(v))
		for i, item := range v {
			cleanedList[i] = sanitizeValue(item)
		}
		return cleanedList
	default:
		return v
	}
}

// sanitizeQuestions ensures answer keys (correct_answer, expected_answer, etc.) NEVER leak to the client.
func sanitizeQuestions(rawQuestions any) any {
	if rawQuestions == nil {
		return []any{}
	}

	rawBytes, err := json.Marshal(rawQuestions)
	if err != nil {
		return []any{}
	}

	var parsed any
	if err := json.Unmarshal(rawBytes, &parsed); err != nil {
		return []any{}
	}

	return sanitizeValue(parsed)
}

func (a *API) HandleGetToday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return
	}
	uid := claims.UID
	familyID := claims.FamilyID

	// 1. Fetch active tasks for today, filtered by family_id
	params := "is_active=eq.true&order=step_order.asc"
	if familyID != "" {
		params = fmt.Sprintf("is_active=eq.true&family_id=eq.%s&order=step_order.asc", familyID)
	}

	taskRaw, err := a.client.Get(ctx, "odyssey_tasks", params)
	if err != nil {
		shared.WriteJSONError(w, "failed to load daily tasks: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var dbTaskList []TaskRecord
	if err := json.Unmarshal(taskRaw, &dbTaskList); err != nil {
		shared.WriteJSONError(w, "invalid tasks payload", http.StatusInternalServerError)
		return
	}

	// In-memory fallback filter if family_id was not handled in URL param
	if familyID != "" {
		filtered := make([]TaskRecord, 0, len(dbTaskList))
		for _, t := range dbTaskList {
			if t.FamilyID == familyID {
				filtered = append(filtered, t)
			}
		}
		dbTaskList = filtered
	}

	// 2. Fetch user's submissions
	subRaw, err := a.client.Get(ctx, "odyssey_task_submissions", fmt.Sprintf("user_uid=eq.%s", uid))
	var submissions []SubmissionRecord
	if err == nil && len(subRaw) > 0 {
		_ = json.Unmarshal(subRaw, &submissions)
	}

	subMap := make(map[int64]SubmissionRecord)
	for _, s := range submissions {
		subMap[s.TaskID] = s
	}

	// 3. Build step views and evaluate linear progression
	// Rule: Step 1 is always unlocked. Step N is unlocked if Step N-1 has been submitted (PENDING or APPROVED).
	taskViews := make([]TaskView, len(dbTaskList))
	var prevStepCompletedOrPending bool = true

	for i, t := range dbTaskList {
		sub, hasSub := subMap[t.ID]

		cfg := t.Config
		if cfg == nil {
			cfg = make(map[string]any)
		} else {
			// Deep copy config to prevent mutation of cache
			cfgCopy := make(map[string]any)
			for k, v := range cfg {
				cfgCopy[k] = v
			}
			cfg = cfgCopy
		}

		// Backward compatibility fields
		if t.YoutubeURL != "" && cfg["youtube_url"] == nil {
			cfg["youtube_url"] = t.YoutubeURL
		}
		if t.Questions != nil && cfg["questions"] == nil {
			cfg["questions"] = t.Questions
		}

		// CRITICAL SECURITY FIX: Sanitize entire config and questions so correct answers NEVER leak to frontend
		if sanitizedCfg, ok := sanitizeValue(cfg).(map[string]any); ok {
			cfg = sanitizedCfg
		}
		if cfg["questions"] != nil {
			cfg["questions"] = sanitizeQuestions(cfg["questions"])
		}

		evalType := tasks.ResolveEvaluationType(t.TaskType, t.EvaluationType)

		view := TaskView{
			ID:             t.ID,
			Title:          t.Title,
			Description:    t.Description,
			TaskType:       t.TaskType,
			EvaluationType: evalType,
			StepOrder:      t.StepOrder,
			RewardCoins:    t.RewardCoins,
			RewardXP:       t.RewardXP,
			Config:         cfg,
		}

		if hasSub {
			view.Status = sub.Status // "APPROVED", "PENDING", or "REJECTED"
			view.CoinsEarned = sub.CoinsEarned
			view.XPEarned = sub.XPEarned
			view.AdminNotes = sub.AdminNotes
			createdAt := sub.CreatedAt
			view.SubmittedAt = &createdAt
			view.IsLocked = false
			// Current step is submitted/done, so next step will be unlocked
			prevStepCompletedOrPending = (sub.Status == "APPROVED" || sub.Status == "PENDING")
		} else {
			if prevStepCompletedOrPending {
				view.Status = "UNLOCKED"
				view.IsLocked = false
			} else {
				view.Status = "LOCKED"
				view.IsLocked = true
			}
			prevStepCompletedOrPending = false
		}

		taskViews[i] = view
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"tasks": taskViews,
	})
}

func (a *API) HandleSubmit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return
	}
	uid := claims.UID

	// Extract Task ID from path: /api/tasks/{id}/submit
	path := strings.TrimPrefix(r.URL.Path, "/api/tasks/")
	path = strings.TrimSuffix(path, "/submit")
	taskID, err := strconv.ParseInt(path, 10, 64)
	if err != nil || taskID <= 0 {
		shared.WriteJSONError(w, "invalid task ID", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		SubmissionType string         `json:"submission_type"` // "AUTO_QUIZ" | "MANUAL_VERIFY"
		Answers        map[string]any `json:"answers,omitempty"`
		Payload        map[string]any `json:"payload,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	// 1. Fetch task details to verify task_type and family_id scoping
	taskRaw, err := a.client.Get(ctx, "odyssey_tasks", fmt.Sprintf("id=eq.%d", taskID))
	if err != nil {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	var taskList []TaskRecord
	if err := json.Unmarshal(taskRaw, &taskList); err != nil || len(taskList) == 0 {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	targetTask := taskList[0]

	// Tenant check: Task must belong to the user's family
	if targetTask.FamilyID != "" && claims.FamilyID != "" && targetTask.FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "access denied: task belongs to another family", http.StatusForbidden)
		return
	}

	// 2. Check if already approved (Anti-double-claim check at API level)
	existingSubRaw, _ := a.client.Get(ctx, "odyssey_task_submissions", fmt.Sprintf("task_id=eq.%d&user_uid=eq.%s", taskID, uid))
	var existingSubs []SubmissionRecord
	if len(existingSubRaw) > 0 {
		_ = json.Unmarshal(existingSubRaw, &existingSubs)
		if len(existingSubs) > 0 && existingSubs[0].Status == "APPROVED" {
			shared.WriteJSONError(w, "task already completed and rewarded", http.StatusBadRequest)
			return
		}
	}

	// 3. Determine if Auto-Graded vs Manual Verification
	resolvedEval := tasks.ResolveEvaluationType(targetTask.TaskType, targetTask.EvaluationType)
	isAuto := resolvedEval == "AUTO" || reqBody.SubmissionType == "AUTO_QUIZ"

	if isAuto && reqBody.SubmissionType != "MANUAL_VERIFY" {
		answers := reqBody.Answers
		if answers == nil {
			answers = reqBody.Payload
		}
		if answers == nil {
			answers = make(map[string]any)
		}

		rpcPayload := map[string]any{
			"p_task_id":  taskID,
			"p_user_uid": uid,
			"p_answers":  answers,
		}
		rpcRes, err := a.client.RPC(ctx, "odyssey_submit_auto_task", rpcPayload)
		if err != nil {
			shared.WriteJSONError(w, "submission failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		var result map[string]any
		_ = json.Unmarshal(rpcRes, &result)
		shared.WriteJSON(w, http.StatusOK, result)
		return
	}

	// Manual verification (PHOTO_UPLOAD, DOCUMENT_UPLOAD, TEXT_RESPONSE, PHOTO_PROOF)
	payload := reqBody.Payload
	if payload == nil {
		payload = reqBody.Answers
	}
	if payload == nil {
		payload = make(map[string]any)
	}

	rpcPayload := map[string]any{
		"p_task_id":  taskID,
		"p_user_uid": uid,
		"p_payload":  payload,
	}
	rpcRes, err := a.client.RPC(ctx, "odyssey_submit_manual_task", rpcPayload)
	if err != nil {
		shared.WriteJSONError(w, "manual submission failed: "+err.Error(), http.StatusBadRequest)
		return
	}
	var result map[string]any
	_ = json.Unmarshal(rpcRes, &result)
	shared.WriteJSON(w, http.StatusOK, result)
}

// Disallowed dangerous extensions
var disallowedExtensions = map[string]bool{
	".exe": true, ".dll": true, ".so": true, ".sh": true,
	".bat": true, ".cmd": true, ".msi": true, ".com": true,
	".vbs": true, ".ps1": true, ".php": true, ".phtml": true,
	".js": true, ".jsp": true, ".asp": true, ".aspx": true,
	".py": true, ".rb": true, ".cgi": true, ".jar": true,
}

// sanitizeFilename strips path traversal sequences and unprintable control characters.
func sanitizeFilename(raw string) string {
	base := filepath.Base(raw)
	base = strings.ReplaceAll(base, "\\", "_")
	base = strings.ReplaceAll(base, "/", "_")
	base = strings.ReplaceAll(base, "..", "")
	base = strings.TrimSpace(base)

	var sb strings.Builder
	for _, r := range base {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') ||
			r == '.' || r == '_' || r == '-' || r == ' ' {
			sb.WriteRune(r)
		} else {
			sb.WriteRune('_')
		}
	}
	res := strings.TrimSpace(sb.String())
	if res == "" || res == "." {
		return "uploaded_file"
	}
	return res
}

func (a *API) HandleUploadProof(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return
	}
	uid := claims.UID
	familyID := claims.FamilyID
	if familyID == "" {
		familyID = "family_default"
	}

	// Max 10MB payload
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		shared.WriteJSONError(w, "file too large (max 10MB)", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		shared.WriteJSONError(w, "file field is required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size <= 0 {
		shared.WriteJSONError(w, "file cannot be empty", http.StatusBadRequest)
		return
	}
	if header.Size > 10<<20 {
		shared.WriteJSONError(w, "file size exceeds 10MB limit", http.StatusBadRequest)
		return
	}

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if disallowedExtensions[ext] {
		shared.WriteJSONError(w, "tipe file executable / script tidak diizinkan", http.StatusBadRequest)
		return
	}

	fileBytes := make([]byte, header.Size)
	if _, err := file.Read(fileBytes); err != nil {
		shared.WriteJSONError(w, "failed to read file: "+err.Error(), http.StatusBadRequest)
		return
	}

	declaredType := header.Header.Get("Content-Type")
	detectedType := http.DetectContentType(fileBytes)
	contentType := declaredType
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = detectedType
	}

	// Reject HTML / script content types to prevent stored XSS (check both declared and detected)
	lowerDeclared := strings.ToLower(declaredType)
	lowerDetected := strings.ToLower(detectedType)
	lowerContent := strings.ToLower(contentType)
	if strings.Contains(lowerContent, "text/html") || strings.Contains(lowerContent, "javascript") || strings.Contains(lowerContent, "application/x-sh") ||
		strings.Contains(lowerDetected, "text/html") || strings.Contains(lowerDetected, "javascript") ||
		strings.Contains(lowerDeclared, "text/html") || strings.Contains(lowerDeclared, "javascript") {
		shared.WriteJSONError(w, "invalid content type", http.StatusBadRequest)
		return
	}

	cleanFileName := sanitizeFilename(header.Filename)

	// Generate secure random nonce for storage path
	randBytes := make([]byte, 4)
	_, _ = rand.Read(randBytes)
	randHex := hex.EncodeToString(randBytes)

	storagePath := fmt.Sprintf("%s/%s/%d_%s_%s", familyID, uid, time.Now().Unix(), randHex, cleanFileName)

	publicURL, err := a.client.UploadStorage(ctx, "task-proofs", storagePath, contentType, fileBytes)
	if err != nil {
		shared.WriteJSONError(w, "failed to upload to storage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"file_url":     publicURL,
		"file_name":    cleanFileName,
		"file_size":    len(fileBytes),
		"storage_path": storagePath,
	})
}

func (a *API) HandleGetTask(w http.ResponseWriter, r *http.Request, taskID int64) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return
	}
	uid := claims.UID
	familyID := claims.FamilyID

	taskRaw, err := a.client.Get(ctx, "odyssey_tasks", fmt.Sprintf("id=eq.%d", taskID))
	if err != nil {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	var taskList []TaskRecord
	if err := json.Unmarshal(taskRaw, &taskList); err != nil || len(taskList) == 0 {
		shared.WriteJSONError(w, "task not found", http.StatusNotFound)
		return
	}
	targetTask := taskList[0]

	// Tenant check: Task must belong to the user's family
	if targetTask.FamilyID != "" && familyID != "" && targetTask.FamilyID != familyID {
		shared.WriteJSONError(w, "access denied: task belongs to another family", http.StatusForbidden)
		return
	}

	// Fetch user's submission for this task if any
	subRaw, _ := a.client.Get(ctx, "odyssey_task_submissions", fmt.Sprintf("task_id=eq.%d&user_uid=eq.%s", taskID, uid))
	var submissions []SubmissionRecord
	if len(subRaw) > 0 {
		_ = json.Unmarshal(subRaw, &submissions)
	}

	cfg := targetTask.Config
	if cfg == nil {
		cfg = make(map[string]any)
	} else {
		cfgCopy := make(map[string]any)
		for k, v := range cfg {
			cfgCopy[k] = v
		}
		cfg = cfgCopy
	}

	if targetTask.YoutubeURL != "" && cfg["youtube_url"] == nil {
		cfg["youtube_url"] = targetTask.YoutubeURL
	}
	if targetTask.Questions != nil && cfg["questions"] == nil {
		cfg["questions"] = targetTask.Questions
	}

	// Sanitize config and questions
	if sanitizedCfg, ok := sanitizeValue(cfg).(map[string]any); ok {
		cfg = sanitizedCfg
	}
	if cfg["questions"] != nil {
		cfg["questions"] = sanitizeQuestions(cfg["questions"])
	}

	evalType := targetTask.EvaluationType
	if evalType == "" {
		switch targetTask.TaskType {
		case "PHOTO_UPLOAD", "DOCUMENT_UPLOAD", "TEXT_RESPONSE", "PHOTO_PROOF":
			evalType = "ADMIN_REVIEW"
		default:
			evalType = "AUTO"
		}
	}

	view := TaskView{
		ID:             targetTask.ID,
		Title:          targetTask.Title,
		Description:    targetTask.Description,
		TaskType:       targetTask.TaskType,
		EvaluationType: evalType,
		StepOrder:      targetTask.StepOrder,
		RewardCoins:    targetTask.RewardCoins,
		RewardXP:       targetTask.RewardXP,
		Config:         cfg,
		Status:         "UNLOCKED",
		IsLocked:       false,
	}

	if len(submissions) > 0 {
		sub := submissions[0]
		view.Status = sub.Status
		view.CoinsEarned = sub.CoinsEarned
		view.XPEarned = sub.XPEarned
		view.AdminNotes = sub.AdminNotes
		createdAt := sub.CreatedAt
		view.SubmittedAt = &createdAt
	}

	shared.WriteJSON(w, http.StatusOK, view)
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	if r.Method == http.MethodGet && path == "/api/tasks/today" {
		a.HandleGetToday(w, r)
		return
	}
	if r.Method == http.MethodPost && path == "/api/tasks/upload" {
		a.HandleUploadProof(w, r)
		return
	}
	if r.Method == http.MethodPost && strings.HasPrefix(path, "/api/tasks/") && strings.HasSuffix(path, "/submit") {
		a.HandleSubmit(w, r)
		return
	}
	if r.Method == http.MethodGet && strings.HasPrefix(path, "/api/tasks/") {
		taskIDStr := strings.TrimPrefix(path, "/api/tasks/")
		if taskID, err := strconv.ParseInt(taskIDStr, 10, 64); err == nil && taskID > 0 {
			a.HandleGetTask(w, r, taskID)
			return
		}
	}
	http.NotFound(w, r)
}
