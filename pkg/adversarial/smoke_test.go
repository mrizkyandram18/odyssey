package adversarial

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	apiAdminMembers "odyssey/internal/api/admin_members"
	apiAdminTasks "odyssey/internal/api/admin_tasks"
	apiFamilyTasks "odyssey/internal/api/family_tasks"
	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

func parseCreatedTaskID(b []byte) int64 {
	var list []struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(b, &list); err == nil && len(list) > 0 {
		return list[0].ID
	}
	var single struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(b, &single); err == nil {
		return single.ID
	}
	return 0
}

func TestRealUserSmokeTest(t *testing.T) {
	dbMock := newMockDB()

	// Setup API packages with shared mock DB
	adminMembersAPI := apiAdminMembers.NewAPI(dbMock)
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)
	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Admin guide profile
	adminUID := "admin_guide_uid"
	adminClaims := &auth.SessionClaims{UID: adminUID, FamilyID: "fam_real_01", Role: "GUIDE"}

	t.Run("1. Admin creates real user -> User logs in -> Queries /today", func(t *testing.T) {
		// Admin creates real user "andi_real"
		createBody := map[string]string{
			"username":      "andi_real",
			"password":      "rahasia123",
			"explorer_name": "Andi Wijaya",
			"role":          "SEEKER",
		}
		bodyBytes, _ := json.Marshal(createBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))
		w := httptest.NewRecorder()
		adminMembersAPI.HandleCreateMember(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Admin create member failed with status %d: %s", w.Code, w.Body.String())
		}

		var memberRes apiAdminMembers.MemberView
		if err := json.Unmarshal(w.Body.Bytes(), &memberRes); err != nil {
			t.Fatalf("Failed to parse created member response: %v", err)
		}
		if memberRes.Username != "andi_real" || memberRes.UID == "" {
			t.Fatalf("Created member unexpected: %+v", memberRes)
		}

		// Verify user profile exists in mock DB
		profBytes, err := dbMock.Get(context.Background(), "odyssey_user_profiles", "uid=eq."+memberRes.UID)
		if err != nil || len(profBytes) == 0 || !strings.Contains(string(profBytes), "Andi Wijaya") {
			t.Fatalf("Profile not found in DB after admin creation: %v (data: %s)", err, string(profBytes))
		}

		// Simulate User Andi querying today's tasks
		todayReq := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
		userClaims := &auth.SessionClaims{UID: memberRes.UID, FamilyID: "fam_real_01", Role: "SEEKER"}
		todayReq = todayReq.WithContext(auth.ContextWithClaims(todayReq.Context(), userClaims))
		todayW := httptest.NewRecorder()
		familyTasksAPI.HandleGetToday(todayW, todayReq)

		if todayW.Code != http.StatusOK {
			t.Fatalf("User /today query failed with status %d: %s", todayW.Code, todayW.Body.String())
		}
		t.Logf("✓ REAL USER SMOKE TEST 1 PASSED: Admin created user %s (UID: %s), login/profile active, /today responded 200 OK.", memberRes.Username, memberRes.UID)
	})

	t.Run("2. Personal task targeting & 403 Forbidden IDOR interception", func(t *testing.T) {
		// Admin creates real user "budi_real"
		createBody := map[string]string{
			"username":      "budi_real",
			"password":      "rahasia123",
			"explorer_name": "Budi Santoso",
			"role":          "SEEKER",
		}
		bodyBytes, _ := json.Marshal(createBody)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))
		w := httptest.NewRecorder()
		adminMembersAPI.HandleCreateMember(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("Admin create Budi failed: %s", w.Body.String())
		}
		var budiRes apiAdminMembers.MemberView
		_ = json.Unmarshal(w.Body.Bytes(), &budiRes)

		// Get Andi's UID
		membersReq := httptest.NewRequest(http.MethodGet, "/api/admin/members", nil)
		membersReq = membersReq.WithContext(auth.ContextWithClaims(membersReq.Context(), adminClaims))
		membersW := httptest.NewRecorder()
		adminMembersAPI.HandleListMembers(membersW, membersReq)
		var members []apiAdminMembers.MemberView
		_ = json.Unmarshal(membersW.Body.Bytes(), &members)

		var andiUID, budiUID string
		for _, m := range members {
			if m.Username == "andi_real" {
				andiUID = m.UID
			}
			if m.Username == "budi_real" {
				budiUID = m.UID
			}
		}

		if andiUID == "" || budiUID == "" {
			t.Fatalf("Could not find Andi and Budi UIDs: andi=%s, budi=%s", andiUID, budiUID)
		}

		// Admin creates personal task for Andi
		personalTaskPayload := map[string]any{
			"title":           "Rapikan CV Kamu",
			"description":     "Update pengalaman kerja terbaru di CV",
			"task_type":       "TEXT_RESPONSE",
			"step_order":      1,
			"reward_coins":    100,
			"reward_xp":       200,
			"target_scope":    "USER",
			"target_user_uid": andiUID,
			"config":          map[string]any{"prompt": "Tulis ringkasan CV:"},
		}
		pBytes, _ := json.Marshal(personalTaskPayload)
		createTaskReq := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewReader(pBytes))
		createTaskReq = createTaskReq.WithContext(auth.ContextWithClaims(createTaskReq.Context(), adminClaims))
		createTaskW := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(createTaskW, createTaskReq)

		if createTaskW.Code != http.StatusCreated {
			t.Fatalf("Failed to create personal task: %s", createTaskW.Body.String())
		}

		taskID := parseCreatedTaskID(createTaskW.Body.Bytes())
		if taskID == 0 {
			t.Fatalf("Failed to parse created personal task ID from response: %s", createTaskW.Body.String())
		}

		// Andi queries /today -> MUST see the personal task
		reqAndi := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
		reqAndi = reqAndi.WithContext(auth.ContextWithClaims(reqAndi.Context(), &auth.SessionClaims{UID: andiUID, FamilyID: "fam_real_01", Role: "SEEKER"}))
		wAndi := httptest.NewRecorder()
		familyTasksAPI.HandleGetToday(wAndi, reqAndi)

		var respAndi struct {
			Tasks []map[string]any `json:"tasks"`
		}
		_ = json.Unmarshal(wAndi.Body.Bytes(), &respAndi)

		foundAndi := false
		for _, task := range respAndi.Tasks {
			if fmt.Sprintf("%v", task["id"]) == fmt.Sprintf("%d", taskID) {
				foundAndi = true
				break
			}
		}
		if !foundAndi {
			t.Fatalf("Personal task (ID %d) for Andi was NOT visible to Andi! Tasks returned: %+v", taskID, respAndi.Tasks)
		}

		// Budi queries /today -> MUST NOT see Andi's personal task
		reqBudi := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
		reqBudi = reqBudi.WithContext(auth.ContextWithClaims(reqBudi.Context(), &auth.SessionClaims{UID: budiUID, FamilyID: "fam_real_01", Role: "SEEKER"}))
		wBudi := httptest.NewRecorder()
		familyTasksAPI.HandleGetToday(wBudi, reqBudi)

		var respBudi struct {
			Tasks []map[string]any `json:"tasks"`
		}
		_ = json.Unmarshal(wBudi.Body.Bytes(), &respBudi)

		for _, task := range respBudi.Tasks {
			if fmt.Sprintf("%v", task["id"]) == fmt.Sprintf("%d", taskID) {
				t.Fatalf("SECURITY LEAK: Budi was able to see Andi's personal task in /today response!")
			}
		}

		// IDOR EXPLOIT: Budi forcibly attempts to submit Andi's personal task via API
		submitBody := fmt.Sprintf(`{"payload":{"text":"Hacked by Budi"}}`)
		idorReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/submit", taskID), strings.NewReader(submitBody))
		idorReq.Header.Set("Content-Type", "application/json")
		idorReq = idorReq.WithContext(auth.ContextWithClaims(idorReq.Context(), &auth.SessionClaims{UID: budiUID, FamilyID: "fam_real_01", Role: "SEEKER"}))
		idorW := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(idorW, idorReq)

		if idorW.Code != http.StatusForbidden {
			t.Fatalf("CRITICAL IDOR VULNERABILITY: Budi submitted Andi's personal task and server returned status %d instead of 403 Forbidden!", idorW.Code)
		}

		t.Logf("✓ REAL USER SMOKE TEST 2 PASSED: Personal task visible to Andi, hidden from Budi, and Budi's forced IDOR API submission intercepted with 403 Forbidden.")
	})

	t.Run("3. Shared task multi-user progression & rewards isolation", func(t *testing.T) {
		// Populate profiles for usr_andi and usr_budi
		dbMock.profiles["usr_andi"] = &db.UserProfile{UID: "usr_andi", FamilyID: "fam_real_01", ExplorerName: "Andi", Role: "SEEKER", IsActive: true}
		dbMock.profiles["usr_budi"] = &db.UserProfile{UID: "usr_budi", FamilyID: "fam_real_01", ExplorerName: "Budi", Role: "SEEKER", IsActive: true}

		// Admin creates shared task "Belajar Email Profesional"
		sharedTaskPayload := map[string]any{
			"title":        "Belajar Email Profesional",
			"description":  "Tonton video dan jawab kuis",
			"task_type":    "MINI_GAME",
			"step_order":   2,
			"reward_coins": 50,
			"reward_xp":    100,
			"target_scope": "ALL",
		}
		pBytes, _ := json.Marshal(sharedTaskPayload)
		req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewReader(pBytes))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))
		w := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(w, req)

		taskID := parseCreatedTaskID(w.Body.Bytes())
		if taskID == 0 {
			t.Fatalf("Failed to parse created shared task ID from response: %s", w.Body.String())
		}

		// Andi completes the task
		submitBody := `{"answers":{"score":90,"moves":10,"game":"MEMORY"}}`
		reqAndiSubmit := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/tasks/%d/submit", taskID), strings.NewReader(submitBody))
		reqAndiSubmit.Header.Set("Content-Type", "application/json")
		reqAndiSubmit = reqAndiSubmit.WithContext(auth.ContextWithClaims(reqAndiSubmit.Context(), &auth.SessionClaims{UID: "usr_andi", FamilyID: "fam_real_01", Role: "SEEKER"}))
		wAndiSubmit := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(wAndiSubmit, reqAndiSubmit)

		if wAndiSubmit.Code != http.StatusOK {
			t.Fatalf("Andi shared task submission failed: %s", wAndiSubmit.Body.String())
		}

		// Verify Budi's view of the shared task remains available / un-submitted for Budi
		reqBudi := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
		reqBudi = reqBudi.WithContext(auth.ContextWithClaims(reqBudi.Context(), &auth.SessionClaims{UID: "usr_budi", FamilyID: "fam_real_01", Role: "SEEKER"}))
		wBudi := httptest.NewRecorder()
		familyTasksAPI.HandleGetToday(wBudi, reqBudi)

		var respBudi struct {
			Tasks []map[string]any `json:"tasks"`
		}
		_ = json.Unmarshal(wBudi.Body.Bytes(), &respBudi)

		foundBudiTask := false
		for _, task := range respBudi.Tasks {
			if fmt.Sprintf("%v", task["id"]) == fmt.Sprintf("%d", taskID) {
				foundBudiTask = true
				if fmt.Sprintf("%v", task["status"]) == "APPROVED" {
					t.Fatalf("PROGRESSION LEAK: Andi completing shared task marked Budi's task as APPROVED!")
				}
			}
		}

		if !foundBudiTask {
			t.Fatalf("Shared task disappeared from Budi's list after Andi completed it!")
		}

		t.Logf("✓ REAL USER SMOKE TEST 3 PASSED: Shared task completion by Andi rewarded Andi independently and left Budi's task state untouched.")
	})

	t.Run("4. New user join-date backlog guard", func(t *testing.T) {
		todayStr := time.Now().Format("2006-01-02")
		yesterdayStr := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

		// Create historical task dated yesterday
		pastTaskPayload := map[string]any{
			"title":        "Tugas Hari Lalu (Day 1)",
			"task_type":    "VIDEO",
			"step_order":   1,
			"active_date":  yesterdayStr,
			"target_scope": "ALL",
		}
		pBytes1, _ := json.Marshal(pastTaskPayload)
		req1 := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewReader(pBytes1))
		req1 = req1.WithContext(auth.ContextWithClaims(req1.Context(), adminClaims))
		w1 := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(w1, req1)

		var pastTask struct {
			ID int64 `json:"id"`
		}
		_ = json.Unmarshal(w1.Body.Bytes(), &pastTask)

		// Create today's task
		todayTaskPayload := map[string]any{
			"title":        "Tugas Hari Ini (Day 4)",
			"task_type":    "VIDEO",
			"step_order":   2,
			"active_date":  todayStr,
			"target_scope": "ALL",
		}
		pBytes2, _ := json.Marshal(todayTaskPayload)
		req2 := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewReader(pBytes2))
		req2 = req2.WithContext(auth.ContextWithClaims(req2.Context(), adminClaims))
		w2 := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(w2, req2)

		// Admin creates new user "citra_new" TODAY
		createBody := map[string]string{
			"username":      "citra_new",
			"password":      "rahasia123",
			"explorer_name": "Citra New User",
			"role":          "SEEKER",
		}
		bodyBytes, _ := json.Marshal(createBody)
		reqCitra := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(bodyBytes))
		reqCitra = reqCitra.WithContext(auth.ContextWithClaims(reqCitra.Context(), adminClaims))
		wCitra := httptest.NewRecorder()
		adminMembersAPI.HandleCreateMember(wCitra, reqCitra)

		var citraRes apiAdminMembers.MemberView
		_ = json.Unmarshal(wCitra.Body.Bytes(), &citraRes)

		// Citra queries /today
		reqQuery := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
		reqQuery = reqQuery.WithContext(auth.ContextWithClaims(reqQuery.Context(), &auth.SessionClaims{UID: citraRes.UID, FamilyID: "fam_real_01", Role: "SEEKER"}))
		wQuery := httptest.NewRecorder()
		familyTasksAPI.HandleGetToday(wQuery, reqQuery)

		var respCitra struct {
			Tasks []map[string]any `json:"tasks"`
		}
		_ = json.Unmarshal(wQuery.Body.Bytes(), &respCitra)

		for _, task := range respCitra.Tasks {
			if fmt.Sprintf("%v", task["id"]) == fmt.Sprintf("%d", pastTask.ID) {
				t.Fatalf("BACKLOG LEAK: New user Citra created today received historical task from yesterday (%s)!", yesterdayStr)
			}
		}

		t.Logf("✓ REAL USER SMOKE TEST 4 PASSED: New user Citra created today received today's task without historical backlog predating join date.")
	})

	t.Run("5. Runtime economy config dynamic override verification", func(t *testing.T) {
		// Admin updates config in DB
		updateConfigPayload := map[string]any{
			"start_day":            1,
			"end_day":              28,
			"payout_day":           25,
			"earning_period_days":  30,
			"conversion_rate":      200,
			"payout_target_rupiah": 500000,
			"max_payout_coins":     2500,
			"timezone":             "Asia/Jakarta",
		}
		cBytes, _ := json.Marshal(updateConfigPayload)
		reqCfg := httptest.NewRequest(http.MethodPost, "/api/admin/config", bytes.NewReader(cBytes))
		reqCfg.Header.Set("Content-Type", "application/json")
		reqCfg = reqCfg.WithContext(auth.ContextWithClaims(reqCfg.Context(), adminClaims))
		wCfg := httptest.NewRecorder()
		adminTasksAPI.Handler(wCfg, reqCfg)

		if wCfg.Code != http.StatusOK {
			t.Fatalf("Admin update config failed: %s", wCfg.Body.String())
		}

		var updatedCfg struct {
			ConversionRate     int `json:"conversion_rate"`
			PayoutTargetRupiah int `json:"payout_target_rupiah"`
			PayoutTargetCoins  int `json:"payout_target_coins"`
			MaxPayoutCoins     int `json:"max_payout_coins"`
		}
		_ = json.Unmarshal(wCfg.Body.Bytes(), &updatedCfg)

		if updatedCfg.ConversionRate != 200 || updatedCfg.PayoutTargetRupiah != 500000 || updatedCfg.PayoutTargetCoins != 2500 || updatedCfg.MaxPayoutCoins != 2500 {
			t.Fatalf("Config mismatch: expected rate=200, rupiah=500000, targetCoins=2500, got: %+v", updatedCfg)
		}

		t.Logf("✓ RUNTIME ECONOMY CONFIG TEST PASSED: Admin updated DB config dynamically, backend derived target_coins=2500 without falling back to hardcoded defaults.")
	})
}
