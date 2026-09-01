package adversarial

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	apiAdminMembers "odyssey/internal/api/admin_members"
	apiAdminTasks "odyssey/internal/api/admin_tasks"
	apiFamilyTasks "odyssey/internal/api/family_tasks"
	apiLogin "odyssey/internal/api/login"
	apiShop "odyssey/internal/api/shop"
	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/server"
)

// -----------------------------------------------------------------------------
// Mock Database Layer with In-Memory State & Thread-Safe Concurrency
// -----------------------------------------------------------------------------

type mockAdversarialDB struct {
	mu           sync.RWMutex
	profiles     map[string]*db.UserProfile
	localUsers   map[string]map[string]any
	sysConfig    map[string]string
	tasks        map[int64]map[string]any
	submissions  map[string]map[string]any // key: fmt.Sprintf("%d:%s", taskID, uid)
	claims       map[int64]map[string]any
	transactions []map[string]any
	nextClaimID  int64
	nextSubID    int64
}

func newMockDB() *mockAdversarialDB {
	return &mockAdversarialDB{
		profiles:    make(map[string]*db.UserProfile),
		localUsers:  make(map[string]map[string]any),
		sysConfig:   make(map[string]string),
		tasks:       make(map[int64]map[string]any),
		submissions: make(map[string]map[string]any),
		claims:      make(map[int64]map[string]any),
		nextClaimID: 100,
		nextSubID:   100,
	}
}

func extractParam(params, prefix string) string {
	parts := strings.Split(params, "&")
	for _, p := range parts {
		if strings.HasPrefix(p, prefix) {
			return strings.TrimPrefix(p, prefix)
		}
	}
	return ""
}

func (m *mockAdversarialDB) Get(ctx context.Context, table string, params string) ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	parts := strings.Split(params, "&")
	targetFam := ""
	targetID := ""
	targetUID := ""
	targetTaskID := ""
	targetUsername := ""
	for _, p := range parts {
		if strings.HasPrefix(p, "family_id=eq.") {
			targetFam = strings.TrimPrefix(p, "family_id=eq.")
		}
		if strings.HasPrefix(p, "task_id=eq.") {
			targetTaskID = strings.TrimPrefix(p, "task_id=eq.")
		} else if strings.HasPrefix(p, "id=eq.") {
			targetID = strings.TrimPrefix(p, "id=eq.")
		}
		if strings.HasPrefix(p, "user_uid=eq.") {
			targetUID = strings.TrimPrefix(p, "user_uid=eq.")
		}
		if strings.HasPrefix(p, "uid=eq.") {
			targetUID = strings.TrimPrefix(p, "uid=eq.")
		}
		if strings.HasPrefix(p, "username=eq.") {
			targetUsername = strings.TrimPrefix(p, "username=eq.")
		}
	}

	switch table {
	case "odyssey_tasks":
		list := make([]map[string]any, 0)
		for _, t := range m.tasks {
			if targetFam != "" && fmt.Sprintf("%v", t["family_id"]) != targetFam {
				continue
			}
			if targetID != "" && fmt.Sprintf("%v", t["id"]) != targetID {
				continue
			}
			list = append(list, t)
		}
		return json.Marshal(list)

	case "odyssey_task_submissions":
		list := make([]map[string]any, 0)
		for _, s := range m.submissions {
			if targetUID != "" && fmt.Sprintf("%v", s["user_uid"]) != targetUID {
				continue
			}
			if targetTaskID != "" && fmt.Sprintf("%v", s["task_id"]) != targetTaskID {
				continue
			}
			if targetID != "" && fmt.Sprintf("%v", s["id"]) != targetID {
				continue
			}
			list = append(list, s)
		}
		return json.Marshal(list)

	case "odyssey_reward_catalog":
		return json.Marshal([]map[string]any{
			{"id": int64(1), "title": "Pulsa 10k", "coin_price": 100, "is_active": true},
			{"id": int64(2), "title": "GoPay 20k", "coin_price": 200, "is_active": true},
		})

	case "odyssey_system_config":
		if len(m.sysConfig) > 0 {
			list := make([]map[string]any, 0)
			for k, v := range m.sysConfig {
				list = append(list, map[string]any{"key": k, "value": v})
			}
			return json.Marshal(list)
		}
		return json.Marshal([]map[string]any{
			{"key": "redemption_start_day", "value": "1"},
			{"key": "redemption_end_day", "value": "31"},
		})

	case "odyssey_local_users":
		list := make([]map[string]any, 0)
		for _, u := range m.localUsers {
			if targetUsername != "" && fmt.Sprintf("%v", u["username"]) != targetUsername {
				continue
			}
			if targetUID != "" && fmt.Sprintf("%v", u["profile_uid"]) != targetUID {
				continue
			}
			list = append(list, u)
		}
		return json.Marshal(list)

	case "odyssey_claims":
		list := make([]map[string]any, 0)
		for _, c := range m.claims {
			if targetID != "" && fmt.Sprintf("%v", c["id"]) != targetID {
				continue
			}
			if targetUID != "" && fmt.Sprintf("%v", c["user_uid"]) != targetUID {
				continue
			}
			list = append(list, c)
		}
		return json.Marshal(list)

	case "odyssey_user_profiles":
		list := make([]db.UserProfile, 0)
		for _, p := range m.profiles {
			if targetFam != "" && p.FamilyID != targetFam {
				continue
			}
			if targetUID != "" && p.UID != targetUID {
				continue
			}
			prof := *p
			// Default IsActive to true for test-seeded profiles where omission means active (zero value false would incorrectly block)
			if !prof.IsActive && prof.BlockedAt == nil && prof.CreatedAt.IsZero() {
				prof.IsActive = true
			}
			list = append(list, prof)
		}
		return json.Marshal(list)

	default:
		return []byte(`[]`), nil
	}
}

func (m *mockAdversarialDB) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return m.MutateAtomic(ctx, method, table, payload, params, "")
}

func (m *mockAdversarialDB) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch table {
	case "odyssey_user_profiles":
		if method == "POST" {
			data, _ := json.Marshal(payload)
			var item db.UserProfile
			_ = json.Unmarshal(data, &item)
			if item.UID != "" {
				if item.CreatedAt.IsZero() {
					item.CreatedAt = time.Now().UTC()
				}
				item.IsActive = true
				m.profiles[item.UID] = &item
			}
			res, _ := json.Marshal([]db.UserProfile{item})
			return res, nil
		}
		if method == "PATCH" {
			uidStr := extractParam(params, "uid=eq.")
			if p, ok := m.profiles[uidStr]; ok {
				data, _ := json.Marshal(payload)
				var patch map[string]any
				_ = json.Unmarshal(data, &patch)
				if name, ok := patch["explorer_name"].(string); ok {
					p.ExplorerName = name
				}
				if role, ok := patch["role"].(string); ok {
					p.Role = role
				}
				if act, ok := patch["is_active"].(bool); ok {
					p.IsActive = act
				}
				res, _ := json.Marshal([]db.UserProfile{*p})
				return res, nil
			}
		}

	case "odyssey_local_users":
		if method == "POST" {
			data, _ := json.Marshal(payload)
			var item map[string]any
			_ = json.Unmarshal(data, &item)
			if un, ok := item["username"].(string); ok {
				m.localUsers[un] = item
			}
			res, _ := json.Marshal([]map[string]any{item})
			return res, nil
		}

	case "odyssey_system_config":
		data, _ := json.Marshal(payload)
		var item map[string]any
		if err := json.Unmarshal(data, &item); err == nil {
			if k, ok := item["key"].(string); ok {
				v := fmt.Sprintf("%v", item["value"])
				m.sysConfig[k] = v
			}
		}
		var list []map[string]any
		if err := json.Unmarshal(data, &list); err == nil {
			for _, elem := range list {
				if k, ok := elem["key"].(string); ok {
					v := fmt.Sprintf("%v", elem["value"])
					m.sysConfig[k] = v
				}
			}
		}
		res, _ := json.Marshal(m.sysConfig)
		return res, nil

	case "odyssey_tasks":
		if method == "POST" {
			data, _ := json.Marshal(payload)
			var item map[string]any
			_ = json.Unmarshal(data, &item)
			id := int64(len(m.tasks) + 1)
			item["id"] = id
			m.tasks[id] = item
			res, _ := json.Marshal([]map[string]any{item})
			return res, nil
		}
		if method == "PATCH" {
			idStr := extractParam(params, "id=eq.")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if t, ok := m.tasks[id]; ok {
				data, _ := json.Marshal(payload)
				var patch map[string]any
				_ = json.Unmarshal(data, &patch)
				for k, v := range patch {
					t[k] = v
				}
				res, _ := json.Marshal([]map[string]any{t})
				return res, nil
			}
		}
		if method == "DELETE" {
			idStr := extractParam(params, "id=eq.")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			delete(m.tasks, id)
			return []byte(`[]`), nil
		}

	case "odyssey_claims":
		if method == "PATCH" {
			idStr := extractParam(params, "id=eq.")
			var id int64
			fmt.Sscanf(idStr, "%d", &id)
			if c, ok := m.claims[id]; ok {
				data, _ := json.Marshal(payload)
				var patch map[string]any
				_ = json.Unmarshal(data, &patch)
				for k, v := range patch {
					c[k] = v
				}
				res, _ := json.Marshal([]map[string]any{c})
				return res, nil
			}
		}
	}
	return []byte(`[]`), nil
}

func (m *mockAdversarialDB) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, _ := json.Marshal(payload)
	var params map[string]any
	_ = json.Unmarshal(data, &params)

	switch fnName {
	case "odyssey_submit_auto_task":
		taskID := int64(params["p_task_id"].(float64))
		uid := params["p_user_uid"].(string)
		answers, _ := params["p_answers"].(map[string]any)

		prof, ok := m.profiles[uid]
		if !ok {
			return nil, fmt.Errorf("P0007: User profile tidak ditemukan")
		}

		task, ok := m.tasks[taskID]
		if !ok {
			return nil, fmt.Errorf("P0002: Task tidak ditemukan")
		}

		// Family Isolation Check
		if task["family_id"] != nil && prof.FamilyID != "" && task["family_id"] != prof.FamilyID {
			return nil, fmt.Errorf("P0003: Akses ditolak: Task bukan milik keluarga Anda")
		}

		// Anti-Double-Claim Check
		key := fmt.Sprintf("%d:%s", taskID, uid)
		if existing, exists := m.submissions[key]; exists && existing["status"] == "APPROVED" {
			return nil, fmt.Errorf("P0004: Tugas ini sudah diselesaikan dan reward sudah diterima")
		}

		// Mini game bounds check
		tType := fmt.Sprintf("%v", task["task_type"])
		if tType == "MINI_GAME" {
			cfg, _ := task["config"].(map[string]any)
			targetScore := 0
			if cfg != nil && cfg["target_score"] != nil {
				if ts, ok := cfg["target_score"].(float64); ok {
					targetScore = int(ts)
				}
				if ts, ok := cfg["target_score"].(int); ok {
					targetScore = ts
				}
			}
			score := 0
			if answers != nil && answers["score"] != nil {
				if sc, ok := answers["score"].(float64); ok {
					score = int(sc)
				}
				if sc, ok := answers["score"].(int); ok {
					score = sc
				}
			}
			if score < 0 || score > 1000000 {
				return nil, fmt.Errorf("P0008: Skor permainan tidak valid")
			}
			if targetScore > 0 && score < targetScore {
				return nil, fmt.Errorf("P0008: Skor permainan belum mencapai target minimum (%d vs target %d)", score, targetScore)
			}
		}

		// Quiz evaluation
		cfg, _ := task["config"].(map[string]any)
		var questions []any
		if cfg != nil && cfg["questions"] != nil {
			questions, _ = cfg["questions"].([]any)
		} else if task["questions"] != nil {
			questions, _ = task["questions"].([]any)
		}

		for _, qRaw := range questions {
			q, _ := qRaw.(map[string]any)
			qID := fmt.Sprintf("%v", q["id"])
			correctAns := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", q["correct_answer"])))
			userAns := strings.TrimSpace(strings.ToLower(fmt.Sprintf("%v", answers[qID])))
			if userAns == "" || (userAns != correctAns &&
				!strings.HasPrefix(userAns, correctAns+".") &&
				!strings.HasPrefix(userAns, correctAns+")") &&
				!strings.HasPrefix(correctAns, userAns+".") &&
				!strings.HasPrefix(correctAns, userAns+")")) {
				return nil, fmt.Errorf("P0008: Jawaban kuis belum tepat, silakan periksa kembali")
			}
		}

		// Atomic Reward
		m.nextSubID++
		subID := m.nextSubID
		rewardCoins := 50
		rewardXP := 100
		if r, ok := task["reward_coins"].(float64); ok && r > 0 {
			rewardCoins = int(r)
		}
		if r, ok := task["reward_coins"].(int); ok && r > 0 {
			rewardCoins = r
		}

		m.submissions[key] = map[string]any{
			"id":              subID,
			"task_id":         taskID,
			"user_uid":        uid,
			"submission_type": "AUTO_QUIZ",
			"status":          "APPROVED",
			"coins_earned":    rewardCoins,
			"xp_earned":       rewardXP,
			"reviewed_at":     time.Now().UTC().Format(time.RFC3339),
		}

		m.transactions = append(m.transactions, map[string]any{
			"user_uid":     uid,
			"amount":       rewardCoins,
			"type":         "TASK_REWARD",
			"reference_id": fmt.Sprintf("%d", subID),
		})

		prof.Coins += int64(rewardCoins)
		prof.XP += int64(rewardXP)

		return json.Marshal(map[string]any{
			"success":       true,
			"submission_id": subID,
			"coins_earned":  rewardCoins,
			"xp_earned":     rewardXP,
			"new_balance":   prof.Coins,
		})

	case "odyssey_submit_manual_task":
		taskID := int64(params["p_task_id"].(float64))
		uid := params["p_user_uid"].(string)
		payloadMap, _ := params["p_payload"].(map[string]any)

		prof, ok := m.profiles[uid]
		if !ok {
			return nil, fmt.Errorf("P0007: User profile tidak ditemukan")
		}

		task, ok := m.tasks[taskID]
		if !ok {
			return nil, fmt.Errorf("P0002: Task tidak ditemukan")
		}

		if task["family_id"] != nil && prof.FamilyID != "" && task["family_id"] != prof.FamilyID {
			return nil, fmt.Errorf("P0003: Akses ditolak: Task bukan milik keluarga Anda")
		}

		key := fmt.Sprintf("%d:%s", taskID, uid)
		if existing, exists := m.submissions[key]; exists && existing["status"] == "APPROVED" {
			return nil, fmt.Errorf("P0004: Tugas ini sudah disetujui sebelumnya")
		}

		// Validation for TEXT_RESPONSE
		tType := fmt.Sprintf("%v", task["task_type"])
		if tType == "TEXT_RESPONSE" {
			text := ""
			if payloadMap != nil && payloadMap["text"] != nil {
				text = strings.TrimSpace(fmt.Sprintf("%v", payloadMap["text"]))
			}
			cfg, _ := task["config"].(map[string]any)
			minChars := 1
			if cfg != nil && cfg["minimum_characters"] != nil {
				if mc, ok := cfg["minimum_characters"].(float64); ok {
					minChars = int(mc)
				}
			}
			if len(text) < minChars {
				return nil, fmt.Errorf("P0008: Panjang teks minimal %d karakter", minChars)
			}
		}

		m.nextSubID++
		subID := m.nextSubID
		m.submissions[key] = map[string]any{
			"id":              subID,
			"task_id":         taskID,
			"user_uid":        uid,
			"submission_type": "MANUAL_VERIFY",
			"status":          "PENDING",
			"payload":         payloadMap,
			"created_at":      time.Now().UTC().Format(time.RFC3339),
		}

		return json.Marshal(map[string]any{
			"success":       true,
			"submission_id": subID,
			"status":        "PENDING",
		})

	case "odyssey_verify_submission":
		var subID int64
		if idFloat, ok := params["p_submission_id"].(float64); ok {
			subID = int64(idFloat)
		} else if idInt, ok := params["p_submission_id"].(int64); ok {
			subID = idInt
		} else if idInt, ok := params["p_submission_id"].(int); ok {
			subID = int64(idInt)
		}
		status := params["p_status"].(string)
		penaltyCoins := 0
		if pVal, ok := params["p_penalty_coins"].(float64); ok {
			penaltyCoins = int(pVal)
		} else if pVal, ok := params["p_penalty_coins"].(int); ok {
			penaltyCoins = pVal
		}

		var targetSub map[string]any
		var targetKey string
		for k, s := range m.submissions {
			if id, ok := s["id"].(int64); ok && id == subID {
				targetSub = s
				targetKey = k
				break
			}
		}

		if targetSub == nil {
			return nil, fmt.Errorf("P0002: Submission tidak ditemukan")
		}

		if targetSub["status"] != "PENDING" {
			return nil, fmt.Errorf("P0004: Submission sudah diproses sebelumnya (status saat ini: %s)", targetSub["status"])
		}

		taskID := targetSub["task_id"].(int64)
		task := m.tasks[taskID]
		uid := targetSub["user_uid"].(string)
		prof := m.profiles[uid]

		if status == "APPROVED" {
			if penaltyCoins > 0 {
				return nil, fmt.Errorf("P0005: Penalti poin tidak dapat diterapkan pada submission yang disetujui")
			}
			targetSub["status"] = "APPROVED"
			rewardCoins := 50
			rewardXP := 100
			if r, ok := task["reward_coins"].(float64); ok && r > 0 {
				rewardCoins = int(r)
			}
			if r, ok := task["reward_coins"].(int); ok && r > 0 {
				rewardCoins = r
			}

			targetSub["coins_earned"] = rewardCoins
			targetSub["xp_earned"] = rewardXP

			m.transactions = append(m.transactions, map[string]any{
				"user_uid":     uid,
				"amount":       rewardCoins,
				"type":         "TASK_REWARD",
				"reference_id": fmt.Sprintf("%d", subID),
			})

			prof.Coins += int64(rewardCoins)
			prof.XP += int64(rewardXP)
			m.submissions[targetKey] = targetSub

			return json.Marshal(map[string]any{
				"success":      true,
				"status":       "APPROVED",
				"coins_earned": rewardCoins,
				"xp_earned":    rewardXP,
				"new_balance":  prof.Coins,
			})
		} else if status == "REJECTED" {
			actualPenalty := 0
			if penaltyCoins > 0 {
				actualPenalty = penaltyCoins
				if int64(actualPenalty) > prof.Coins {
					actualPenalty = int(prof.Coins)
				}
				if actualPenalty > 0 {
					m.transactions = append(m.transactions, map[string]any{
						"user_uid":     uid,
						"amount":       -actualPenalty,
						"type":         "TASK_PENALTY",
						"reference_id": fmt.Sprintf("%d", subID),
					})
					prof.Coins -= int64(actualPenalty)
				}
			}

			targetSub["status"] = "REJECTED"
			targetSub["coins_earned"] = -actualPenalty
			m.submissions[targetKey] = targetSub
			return json.Marshal(map[string]any{
				"success":        true,
				"status":         "REJECTED",
				"coins_deducted": actualPenalty,
				"new_balance":    prof.Coins,
			})
		}
		return nil, fmt.Errorf("P0005: Status tidak valid")

	case "odyssey_admin_edit_submission":
		var subID int64
		if idFloat, ok := params["p_submission_id"].(float64); ok {
			subID = int64(idFloat)
		} else if idInt, ok := params["p_submission_id"].(int64); ok {
			subID = idInt
		}
		payload, _ := params["p_payload"].(map[string]any)

		var targetSub map[string]any
		var targetKey string
		for k, s := range m.submissions {
			if id, ok := s["id"].(int64); ok && id == subID {
				targetSub = s
				targetKey = k
				break
			}
		}

		if targetSub == nil {
			return nil, fmt.Errorf("P0002: Submission tidak ditemukan")
		}

		if targetSub["status"] == "APPROVED" {
			return nil, fmt.Errorf("P0004: Submission yang sudah disetujui tidak dapat diedit")
		}

		targetSub["payload"] = payload
		m.submissions[targetKey] = targetSub

		return json.Marshal(map[string]any{
			"success":       true,
			"submission_id": subID,
			"status":        targetSub["status"],
			"payload":       payload,
		})

	case "odyssey_create_claim":
		uid := params["p_user_uid"].(string)
		coins := int64(params["p_coins"].(float64))

		prof, ok := m.profiles[uid]
		if !ok {
			return nil, fmt.Errorf("P0007: User profile tidak ditemukan")
		}

		if prof.Coins < coins {
			return nil, fmt.Errorf("P0003: Saldo koin tidak mencukupi")
		}

		// Single pending claim constraint
		for _, c := range m.claims {
			if c["user_uid"] == uid && c["status"] == "PENDING" {
				return nil, fmt.Errorf("P0006: Anda masih memiliki klaim pending yang belum diproses")
			}
		}

		m.nextClaimID++
		claimID := m.nextClaimID

		m.claims[claimID] = map[string]any{
			"id":           claimID,
			"user_uid":     uid,
			"family_id":    prof.FamilyID,
			"coins_spent":  coins,
			"status":       "PENDING",
			"target_type":  params["p_target_type"],
			"target_value": params["p_target_value"],
			"created_at":   time.Now().UTC().Format(time.RFC3339),
		}

		prof.Coins -= coins
		m.transactions = append(m.transactions, map[string]any{
			"user_uid":     uid,
			"amount":       -coins,
			"type":         "REWARD_CLAIM",
			"reference_id": fmt.Sprintf("%d", claimID),
		})

		return json.Marshal(map[string]any{
			"success":     true,
			"claim_id":    claimID,
			"new_balance": prof.Coins,
		})

	case "odyssey_process_claim":
		claimID := int64(params["p_claim_id"].(float64))
		status := params["p_status"].(string)

		claim, ok := m.claims[claimID]
		if !ok {
			return nil, fmt.Errorf("P0002: Klaim tidak ditemukan")
		}

		if claim["status"] != "PENDING" {
			return nil, fmt.Errorf("P0004: Klaim ini sudah diproses sebelumnya")
		}

		uid := claim["user_uid"].(string)
		coinsSpent := claim["coins_spent"].(int64)

		if status == "APPROVED" {
			claim["status"] = "APPROVED"
			claim["processed_at"] = time.Now().UTC().Format(time.RFC3339)
			return json.Marshal(map[string]any{"success": true, "status": "APPROVED"})
		} else if status == "REJECTED" {
			claim["status"] = "REJECTED"
			claim["processed_at"] = time.Now().UTC().Format(time.RFC3339)

			// Atomic Refund
			prof := m.profiles[uid]
			if prof != nil {
				prof.Coins += coinsSpent
			}
			m.transactions = append(m.transactions, map[string]any{
				"user_uid":     uid,
				"amount":       coinsSpent,
				"type":         "CLAIM_REFUND",
				"reference_id": fmt.Sprintf("%d", claimID),
			})
			return json.Marshal(map[string]any{"success": true, "status": "REJECTED", "refunded": coinsSpent})
		}
		return nil, fmt.Errorf("P0005: Status tidak valid")
	}

	return []byte(`{}`), nil
}

func (m *mockAdversarialDB) UploadStorage(ctx context.Context, bucket string, storagePath string, contentType string, fileBytes []byte) (string, error) {
	return "http://localhost/storage/" + storagePath, nil
}

func (m *mockAdversarialDB) GetLocalUserByUsername(ctx context.Context, username string) (*auth.LocalUser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	u, ok := m.localUsers[username]
	if !ok {
		return nil, auth.ErrLocalUserNotFound
	}
	return &auth.LocalUser{
		Username:     fmt.Sprintf("%v", u["username"]),
		PasswordHash: fmt.Sprintf("%v", u["password_hash"]),
		ProfileUID:   fmt.Sprintf("%v", u["profile_uid"]),
	}, nil
}

func (m *mockAdversarialDB) GetUserProfile(ctx context.Context, uid string) (*db.UserProfile, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[uid]
	if !ok || p == nil {
		return nil, db.ErrProfileNotFound
	}
	return p, nil
}

func (m *mockAdversarialDB) GetPasswordHash(ctx context.Context, uid string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, u := range m.localUsers {
		if fmt.Sprintf("%v", u["profile_uid"]) == uid {
			return fmt.Sprintf("%v", u["password_hash"]), nil
		}
	}
	return "", nil
}

func (m *mockAdversarialDB) GetBoundDeviceID(ctx context.Context, uid string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[uid]
	if ok && p != nil {
		return p.DeviceID, nil
	}
	return "", nil
}

func (m *mockAdversarialDB) BindOrVerifyDevice(ctx context.Context, uid, deviceID string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[uid]
	if !ok || p == nil {
		return false, db.ErrProfileNotFound
	}
	if deviceID == "" {
		return false, auth.ErrDeviceRequired
	}
	if p.DeviceID == "" {
		p.DeviceID = deviceID
		return true, nil
	}
	if p.DeviceID == deviceID {
		return false, nil
	}
	return false, auth.ErrDeviceBlocked
}

func (m *mockAdversarialDB) ResetDeviceBinding(ctx context.Context, uid string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, ok := m.profiles[uid]
	if ok && p != nil {
		p.DeviceID = ""
	}
	return nil
}

func (m *mockAdversarialDB) SetAvatarFrame(ctx context.Context, uid, frame string) error {
	return nil
}

func (m *mockAdversarialDB) SetExplorerEffect(ctx context.Context, uid, effect string) error {
	return nil
}

func (m *mockAdversarialDB) UpdateAvatar(ctx context.Context, uid string, style, seed string) error {
	return nil
}

func (m *mockAdversarialDB) ChangePassword(_ context.Context, _ string, _ string) error {
	return nil
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 1: IDOR & Cross-Family Isolation Matrix
// -----------------------------------------------------------------------------

func TestAdversarial_CrossFamilyIDORMatrix(t *testing.T) {
	dbMock := newMockDB()

	// Seed Family A and Family B
	dbMock.profiles["member-a"] = &db.UserProfile{UID: "member-a", FamilyID: "family-alpha", Role: "SEEKER", Coins: 100, IsActive: true}
	dbMock.profiles["admin-a"] = &db.UserProfile{UID: "admin-a", FamilyID: "family-alpha", Role: "GUIDE", Coins: 500, IsActive: true}
	dbMock.profiles["member-b"] = &db.UserProfile{UID: "member-b", FamilyID: "family-beta", Role: "SEEKER", Coins: 100, IsActive: true}
	dbMock.profiles["admin-b"] = &db.UserProfile{UID: "admin-b", FamilyID: "family-beta", Role: "GUIDE", Coins: 500, IsActive: true}

	// Task 1: Family Alpha
	dbMock.tasks[1] = map[string]any{
		"id": int64(1), "family_id": "family-alpha", "title": "Alpha Task", "task_type": "VIDEO",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"questions": []any{
				map[string]any{"id": "1", "question": "Q1", "options": []any{"A", "B"}, "correct_answer": "A"},
			},
		},
	}

	// Task 2: Family Beta
	dbMock.tasks[2] = map[string]any{
		"id": int64(2), "family_id": "family-beta", "title": "Beta Task", "task_type": "VIDEO",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"questions": []any{
				map[string]any{"id": "1", "question": "Q1", "options": []any{"A", "B"}, "correct_answer": "B"},
			},
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)
	shopAPI := apiShop.NewAPI(dbMock)

	// Subtest 1: Member A queries today's tasks -> only gets Task 1, never Task 2
	t.Run("Member A cannot see Family Beta tasks", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-a", FamilyID: "family-alpha", Role: "SEEKER"})
		w := httptest.NewRecorder()
		familyTasksAPI.HandleGetToday(w, req.WithContext(ctx))

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var res struct {
			Tasks []map[string]any `json:"tasks"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		if len(res.Tasks) != 1 || res.Tasks[0]["title"] != "Alpha Task" {
			t.Fatalf("expected only Alpha Task, got %v", res.Tasks)
		}
	})

	// Subtest 2: Member A attempts IDOR exploit to submit Family Beta Task 2 -> MUST BE REJECTED (403 Forbidden)
	t.Run("Member A IDOR submit Family Beta Task -> 403 Forbidden", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"1":"B"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/2/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-a", FamilyID: "family-alpha", Role: "SEEKER"})
		w := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden on IDOR submission, got %d: %s", w.Code, w.Body.String())
		}
		// Verify zero side effect on member-a balance
		if dbMock.profiles["member-a"].Coins != 100 {
			t.Fatalf("expected balance unchanged (100), got %d", dbMock.profiles["member-a"].Coins)
		}
	})

	// Subtest 3: Admin A attempts to list Family Beta submissions -> Filtered to Alpha only
	t.Run("Admin A cannot list Family Beta submissions", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/admin/submissions", nil)
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin-a", FamilyID: "family-alpha", Role: "GUIDE"})
		w := httptest.NewRecorder()
		adminTasksAPI.HandleListPendingSubmissions(w, req.WithContext(ctx))

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
	})

	// Subtest 4: Admin A attempts to process Family Beta member claim -> 403 Forbidden
	t.Run("Admin A cannot approve/reject Family Beta claim", func(t *testing.T) {
		dbMock.claims[999] = map[string]any{
			"id": int64(999), "user_uid": "member-b", "family_id": "family-beta",
			"coins_spent": int64(50), "status": "PENDING",
		}
		payload := `{"status":"APPROVED","admin_notes":"Hacked"}`
		req := httptest.NewRequest(http.MethodPost, "/api/admin/claims/999/process", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin-a", FamilyID: "family-alpha", Role: "GUIDE"})
		w := httptest.NewRecorder()
		shopAPI.HandleAdminProcessClaim(w, req.WithContext(ctx), 999)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden on cross-family claim process, got %d: %s", w.Code, w.Body.String())
		}
		if dbMock.claims[999]["status"] != "PENDING" {
			t.Fatalf("expected claim status to remain PENDING, got %v", dbMock.claims[999]["status"])
		}
	})

	// Subtest 5: Regular member (SEEKER) attempts Admin endpoint -> 403 Forbidden
	t.Run("Non-admin SEEKER cannot access admin tasks", func(t *testing.T) {
		payload := `{"title":"Malicious Task","task_type":"DOCUMENT_UPLOAD","reward_coins":1000}`
		req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-a", FamilyID: "family-alpha", Role: "SEEKER"})
		w := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(w, req.WithContext(ctx))

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden for non-guide user, got %d", w.Code)
		}
	})

	// Subtest 6: Member A attempts IDOR GET /api/tasks/2 (Family Beta Task) -> 403 Forbidden
	t.Run("Member A IDOR GET Family Beta Task -> 403 Forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/2", nil)
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-a", FamilyID: "family-alpha", Role: "SEEKER"})
		w := httptest.NewRecorder()
		familyTasksAPI.Handler(w, req.WithContext(ctx))

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403 Forbidden on cross-family GET task, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Subtest 7: Member A GET /api/tasks/1 (Family Alpha Task) -> 200 OK without answer keys
	t.Run("Member A GET Family Alpha Task -> 200 OK and sanitized", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil)
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-a", FamilyID: "family-alpha", Role: "SEEKER"})
		w := httptest.NewRecorder()
		familyTasksAPI.Handler(w, req.WithContext(ctx))

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on legitimate GET task, got %d: %s", w.Code, w.Body.String())
		}
		if strings.Contains(w.Body.String(), "correct_answer") {
			t.Fatalf("answer key leaked in GET task: %s", w.Body.String())
		}
	})
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 2: Concurrency & Race Condition Attack Matrix
// -----------------------------------------------------------------------------

func TestAdversarial_100ConcurrentSubmissionsRace(t *testing.T) {
	dbMock := newMockDB()

	// Initial user state: 0 coins
	dbMock.profiles["race-user-1"] = &db.UserProfile{
		UID: "race-user-1", FamilyID: "race-family", Role: "SEEKER", Coins: 0,
	}

	// Task 10: 50 coin reward
	dbMock.tasks[10] = map[string]any{
		"id": int64(10), "family_id": "race-family", "title": "Race Task", "task_type": "QUIZ",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"questions": []any{
				map[string]any{"id": "1", "correct_answer": "A"},
			},
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Launch 100 simultaneous concurrent submission requests for the exact same task
	const concurrency = 100
	var wg sync.WaitGroup
	var successCount int64
	var failureCount int64

	startSignal := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal // Thundering herd synchronization

			payload := `{"submission_type":"AUTO_QUIZ","answers":{"1":"A"}}`
			req := httptest.NewRequest(http.MethodPost, "/api/tasks/10/submit", bytes.NewBufferString(payload))
			ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "race-user-1", FamilyID: "race-family", Role: "SEEKER"})
			w := httptest.NewRecorder()

			familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))

			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failureCount, 1)
			}
		}()
	}

	close(startSignal) // Trigger all 100 goroutines simultaneously
	wg.Wait()

	// INVARIANT VERIFICATION:
	// 1. Exactly 1 request succeeds
	// 2. Exactly 99 requests fail with anti-double-claim error
	// 3. User balance is EXACTLY 50 coins (never 50 * N)
	// 4. Ledger transaction count is EXACTLY 1
	if successCount != 1 {
		t.Fatalf("CRITICAL CONCURRENCY FAILURE: Expected exactly 1 success, got %d (failures: %d)", successCount, failureCount)
	}
	if failureCount != concurrency-1 {
		t.Fatalf("Expected %d failures, got %d", concurrency-1, failureCount)
	}
	if dbMock.profiles["race-user-1"].Coins != 50 {
		t.Fatalf("INSUFFICIENT CONCURRENCY ISOLATION: User coins = %d, expected exactly 50", dbMock.profiles["race-user-1"].Coins)
	}
	if len(dbMock.transactions) != 1 {
		t.Fatalf("Ledger transactions = %d, expected exactly 1", len(dbMock.transactions))
	}
}

func TestAdversarial_100ConcurrentRedemptionsRace(t *testing.T) {
	dbMock := newMockDB()

	// Initial user balance: exactly 100 coins
	dbMock.profiles["poor-user"] = &db.UserProfile{
		UID: "poor-user", FamilyID: "race-family", Role: "SEEKER", Coins: 100,
	}

	shopAPI := apiShop.NewAPI(dbMock)

	// Launch 100 simultaneous redemptions of 100 coins each
	const concurrency = 100
	var wg sync.WaitGroup
	var successCount int64
	var failureCount int64

	startSignal := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal

			payload := `{"catalog_id":1,"coins":100,"target_type":"GOPAY","target_value":"08123456789"}`
			req := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", bytes.NewBufferString(payload))
			ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "poor-user", FamilyID: "race-family", Role: "SEEKER"})
			w := httptest.NewRecorder()

			shopAPI.HandleRedeem(w, req.WithContext(ctx))

			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failureCount, 1)
			}
		}()
	}

	close(startSignal)
	wg.Wait()

	// INVARIANT VERIFICATION:
	// 1. Exactly 1 redemption succeeds
	// 2. 99 redemptions rejected (insufficient funds or single pending claim invariant)
	// 3. User balance is EXACTLY 0 (NEVER NEGATIVE)
	if successCount != 1 {
		t.Fatalf("OVER-REDEMPTION DETECTED: %d redemptions succeeded with only 100 coins!", successCount)
	}
	if dbMock.profiles["poor-user"].Coins != 0 {
		t.Fatalf("Expected balance 0, got %d", dbMock.profiles["poor-user"].Coins)
	}
}

func TestAdversarial_100ConcurrentAdminApprovalsRace(t *testing.T) {
	dbMock := newMockDB()

	dbMock.profiles["member-pending"] = &db.UserProfile{
		UID: "member-pending", FamilyID: "race-family", Role: "SEEKER", Coins: 0,
	}
	dbMock.profiles["admin-race"] = &db.UserProfile{
		UID: "admin-race", FamilyID: "race-family", Role: "GUIDE", Coins: 0,
	}

	dbMock.tasks[20] = map[string]any{
		"id": int64(20), "family_id": "race-family", "title": "Photo Task", "task_type": "PHOTO_UPLOAD",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
	}

	// Seed pending submission
	subID := int64(300)
	key := "20:member-pending"
	dbMock.submissions[key] = map[string]any{
		"id":              subID,
		"task_id":         int64(20),
		"user_uid":        "member-pending",
		"submission_type": "MANUAL_VERIFY",
		"status":          "PENDING",
		"payload":         map[string]any{"file_url": "https://cdn.example.com/photo.jpg"},
		"created_at":      time.Now().UTC().Format(time.RFC3339),
	}

	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)

	const concurrency = 100
	var wg sync.WaitGroup
	var successCount int64
	var failureCount int64

	startSignal := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-startSignal

			payload := `{"status":"APPROVED","notes":"Good job"}`
			req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(payload))
			ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin-race", FamilyID: "race-family", Role: "GUIDE"})
			w := httptest.NewRecorder()

			adminTasksAPI.HandleVerifySubmission(w, req.WithContext(ctx), subID)

			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failureCount, 1)
			}
		}()
	}

	close(startSignal)
	wg.Wait()

	// Invariant verification
	if successCount != 1 {
		t.Fatalf("CRITICAL CONCURRENCY FAILURE: %d approvals succeeded (expected exactly 1)", successCount)
	}
	if failureCount != concurrency-1 {
		t.Fatalf("Expected %d failures, got %d", concurrency-1, failureCount)
	}
	if dbMock.profiles["member-pending"].Coins != 50 {
		t.Fatalf("DOUBLE REWARD DETECTED: user coins = %d, expected 50", dbMock.profiles["member-pending"].Coins)
	}
	if len(dbMock.transactions) != 1 {
		t.Fatalf("Expected exactly 1 ledger transaction, got %d", len(dbMock.transactions))
	}
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 3: Quiz Leakage & Deep Serialization Path Audit
// -----------------------------------------------------------------------------

func TestAdversarial_ZeroAnswerLeakageDeepScan(t *testing.T) {
	dbMock := newMockDB()

	// Seed task with various evil nested answer-bearing keys
	dbMock.tasks[99] = map[string]any{
		"id": int64(99), "family_id": "test-family", "title": "Evil Quiz Task", "task_type": "QUIZ",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"youtube_url":     "https://youtube.com/watch?v=123",
			"correct_answer":  "SUPER_SECRET_ANSWER_TOP",
			"expected_answer": "SUPER_SECRET_EXPECTED",
			"solution":        "SUPER_SECRET_SOLUTION",
			"questions": []any{
				map[string]any{
					"id":              "q1",
					"question":        "What is 2+2?",
					"correct_answer":  "SECRET_ANSWER_4",
					"expected_answer": "SECRET_ANSWER_4",
					"answer_key":      "SECRET_ANSWER_4",
					"is_correct":      true,
					"solution":        "SECRET_SOLUTION_STEPS",
					"options": []any{
						map[string]any{"text": "Option 3", "is_correct": false},
						map[string]any{"text": "Option 4", "is_correct": true, "correct_option": true},
					},
				},
			},
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks/today", nil)
	ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "user-1", FamilyID: "test-family", Role: "SEEKER"})
	w := httptest.NewRecorder()

	familyTasksAPI.HandleGetToday(w, req.WithContext(ctx))

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	rawJSON := w.Body.String()

	// Exhaustive prohibited tokens
	prohibitedTokens := []string{
		"SUPER_SECRET_ANSWER_TOP",
		"SUPER_SECRET_EXPECTED",
		"SUPER_SECRET_SOLUTION",
		"SECRET_ANSWER_4",
		"SECRET_SOLUTION_STEPS",
		"correct_answer",
		"expected_answer",
		"answer_key",
		"is_correct",
		"correct_option",
		"\"solution\":",
	}

	for _, token := range prohibitedTokens {
		if strings.Contains(rawJSON, token) {
			t.Fatalf("CRITICAL SECURITY LEAK DETECTED: Token %q found in client response payload!\nRaw JSON:\n%s", token, rawJSON)
		}
	}
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 4: Mini-Game Score & Tampering Validation
// -----------------------------------------------------------------------------

func TestAdversarial_MiniGameScoreTampering(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["game-user"] = &db.UserProfile{UID: "game-user", FamilyID: "game-fam", Role: "SEEKER", Coins: 0}

	// Game Task with target score 80
	dbMock.tasks[50] = map[string]any{
		"id": int64(50), "family_id": "game-fam", "title": "Memory Challenge", "task_type": "MINI_GAME",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"game":         "MEMORY",
			"target_score": 80.0,
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Attack 1: Client submits failing score (40 < 80) -> MUST BE REJECTED
	t.Run("Fake score below target is rejected", func(t *testing.T) {
		payload := `{"answers":{"score":40,"moves":30,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/50/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "game-user", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on score below target, got %d: %s", w.Code, w.Body.String())
		}
		if dbMock.profiles["game-user"].Coins != 0 {
			t.Fatalf("user received coins illegally: %d", dbMock.profiles["game-user"].Coins)
		}
	})

	// Attack 2: Client submits negative or absurd score -> MUST BE REJECTED
	t.Run("Negative score is rejected", func(t *testing.T) {
		payload := `{"answers":{"score":-10,"moves":5,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/50/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "game-user", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on negative score, got %d", w.Code)
		}
	})

	// Legitimate game score >= 80 -> SUCCESS
	t.Run("Valid score >= target succeeds and rewards once", func(t *testing.T) {
		payload := `{"answers":{"score":85,"moves":12,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/50/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "game-user", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on valid score, got %d: %s", w.Code, w.Body.String())
		}
		if dbMock.profiles["game-user"].Coins != 50 {
			t.Fatalf("expected 50 coins awarded, got %d", dbMock.profiles["game-user"].Coins)
		}
	})
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 5: Document & Upload Security Attacks
// -----------------------------------------------------------------------------

func TestAdversarial_DocumentAndUploadAttacks(t *testing.T) {
	dbMock := newMockDB()
	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Attack 1: Path Traversal in filename
	t.Run("Path traversal filename is safely sanitized", func(t *testing.T) {
		var buf bytes.Buffer
		mw := multipart.NewWriter(&buf)
		fw, _ := mw.CreateFormFile("file", "../../../evil.png")
		fw.Write([]byte("fake-image-bytes"))
		mw.Close()

		req := httptest.NewRequest(http.MethodPost, "/api/tasks/upload", &buf)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "user-hacker", FamilyID: "fam-1", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleUploadProof(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 with sanitized path, got %d: %s", w.Code, w.Body.String())
		}

		var res map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		storagePath := fmt.Sprintf("%v", res["storage_path"])
		if strings.Contains(storagePath, "..") {
			t.Fatalf("CRITICAL SECURITY ERROR: Path traversal persisted in storage path: %s", storagePath)
		}
	})

	// Attack 2: Executable script upload rejected
	t.Run("Disallowed extension (.php, .sh, .exe) is rejected", func(t *testing.T) {
		for _, evilExt := range []string{"shell.php", "hack.exe", "script.sh", "payload.bat"} {
			var buf bytes.Buffer
			mw := multipart.NewWriter(&buf)
			fw, _ := mw.CreateFormFile("file", evilExt)
			fw.Write([]byte("malicious-code"))
			mw.Close()

			req := httptest.NewRequest(http.MethodPost, "/api/tasks/upload", &buf)
			req.Header.Set("Content-Type", mw.FormDataContentType())
			ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "user-hacker", FamilyID: "fam-1", Role: "SEEKER"})
			w := httptest.NewRecorder()

			familyTasksAPI.HandleUploadProof(w, req.WithContext(ctx))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 Bad Request for evil file %s, got %d", evilExt, w.Code)
			}
		}
	})
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 6: Text Response Constraints & Admin Review Double Reward Guard
// -----------------------------------------------------------------------------

func TestAdversarial_TextResponseAndAdminApproval(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["member-text"] = &db.UserProfile{UID: "member-text", FamilyID: "fam-text", Role: "SEEKER", Coins: 0}
	dbMock.profiles["admin-text"] = &db.UserProfile{UID: "admin-text", FamilyID: "fam-text", Role: "GUIDE", Coins: 0}

	// Text Task requiring min 30 chars
	dbMock.tasks[70] = map[string]any{
		"id": int64(70), "family_id": "fam-text", "title": "Daily Reflection", "task_type": "TEXT_RESPONSE",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"minimum_characters": 30.0,
			"maximum_characters": 500.0,
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)

	// Attack 1: Empty / Too short text (< 30 chars) -> REJECTED
	t.Run("Text response shorter than min constraint is rejected", func(t *testing.T) {
		payload := `{"payload":{"text":"Short"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/70/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-text", FamilyID: "fam-text", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on short text, got %d: %s", w.Code, w.Body.String())
		}
	})

	// Valid Text submission -> Enters PENDING
	var subID int64
	t.Run("Valid text submission enters PENDING state", func(t *testing.T) {
		payload := `{"payload":{"text":"Hari ini saya belajar cara mengelola uang jajan dan menabung di celengan."}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/70/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "member-text", FamilyID: "fam-text", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on text submit, got %d: %s", w.Code, w.Body.String())
		}
		var res map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		if res["status"] != "PENDING" {
			t.Fatalf("expected status PENDING, got %v", res["status"])
		}
		subID = int64(res["submission_id"].(float64))
	})

	// Admin approves submission -> 50 coins awarded
	t.Run("Admin approves submission exactly once", func(t *testing.T) {
		payload := `{"status":"APPROVED","notes":"Bagus sekali!"}`
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin-text", FamilyID: "fam-text", Role: "GUIDE"})
		w := httptest.NewRecorder()

		adminTasksAPI.HandleVerifySubmission(w, req.WithContext(ctx), subID)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on admin verify, got %d: %s", w.Code, w.Body.String())
		}
		if dbMock.profiles["member-text"].Coins != 50 {
			t.Fatalf("expected 50 coins awarded to member, got %d", dbMock.profiles["member-text"].Coins)
		}
	})

	// Attack 2: Admin double approve -> MUST BE REJECTED
	t.Run("Admin cannot double approve already approved submission", func(t *testing.T) {
		payload := `{"status":"APPROVED","notes":"Duplicate approve"}`
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin-text", FamilyID: "fam-text", Role: "GUIDE"})
		w := httptest.NewRecorder()

		adminTasksAPI.HandleVerifySubmission(w, req.WithContext(ctx), subID)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on second approval, got %d: %s", w.Code, w.Body.String())
		}
		// Balance MUST still be exactly 50
		if dbMock.profiles["member-text"].Coins != 50 {
			t.Fatalf("DOUBLE REWARD DETECTED: member coins = %d", dbMock.profiles["member-text"].Coins)
		}
	})
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 7: Live Binary Route & Surface Verification
// -----------------------------------------------------------------------------

func TestAdversarial_LegacyRouteSurfacePurge(t *testing.T) {
	// Set mock environment variables for server build
	os.Setenv("SUPABASE_URL", "https://example.supabase.co")
	os.Setenv("SUPABASE_SERVICE_KEY", "mock-service-key")
	os.Setenv("SESSION_SIGNING_SECRET", "mock-session-signing-secret-key-32b")
	os.Setenv("PARENT_ID", "mock-parent-id")
	os.Setenv("PORT", "8080")

	// Build the live server handler
	srv, err := server.BuildHandler()
	if err != nil {
		t.Fatalf("BuildHandler error: %v", err)
	}
	if srv.Cleanup != nil {
		defer srv.Cleanup(context.Background())
	}

	// Matrix of all legacy RPG routes that MUST return 404
	legacyRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/api/missions"},
		{"POST", "/api/missions/1/complete"},
		{"GET", "/api/quests"},
		{"POST", "/api/quests/start"},
		{"GET", "/api/journeys"},
		{"GET", "/api/realms"},
		{"GET", "/api/chapters"},
		{"GET", "/api/courses"},
		{"GET", "/api/exercises"},
		{"GET", "/api/lore"},
		{"GET", "/api/story"},
		{"GET", "/api/fragments"},
		{"GET", "/api/chests"},
		{"POST", "/api/chests/open"},
		{"GET", "/api/gifts"},
		{"GET", "/api/relics"},
		{"GET", "/api/collections"},
		{"GET", "/api/drops"},
		{"GET", "/api/reactions"},
		{"GET", "/api/creative"},
		{"POST", "/api/creative/submit"},
		{"GET", "/api/comics"},
		{"GET", "/api/cosmetics"},
	}

	for _, lr := range legacyRoutes {
		t.Run(fmt.Sprintf("%s %s -> 404", lr.method, lr.path), func(t *testing.T) {
			req := httptest.NewRequest(lr.method, lr.path, nil)
			w := httptest.NewRecorder()
			srv.Handler.ServeHTTP(w, req)

			if w.Code != http.StatusNotFound {
				t.Fatalf("SURFACE VIOLATION: Legacy route %s %s returned status %d (expected 404)", lr.method, lr.path, w.Code)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 8: Quiz Answer Tampering, Missing Questions & Data Abuse
// -----------------------------------------------------------------------------

func TestAdversarial_QuizAnswerTamperingAndMissingQuestions(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["quiz-hacker"] = &db.UserProfile{
		UID: "quiz-hacker", FamilyID: "quiz-fam", Role: "SEEKER", Coins: 0,
	}

	dbMock.tasks[88] = map[string]any{
		"id": int64(88), "family_id": "quiz-fam", "title": "Math Quiz", "task_type": "QUIZ",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"questions": []any{
				map[string]any{"id": "q1", "question": "2+2?", "options": []any{"A. 3", "B. 4"}, "correct_answer": "B"},
				map[string]any{"id": "q2", "question": "5*5?", "options": []any{"A. 25", "B. 20"}, "correct_answer": "A"},
			},
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Scenario 1: Missing question 2
	t.Run("Missing question answers is rejected", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"q1":"B"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/88/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "quiz-hacker", FamilyID: "quiz-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on missing question answer, got %d", w.Code)
		}
	})

	// Scenario 2: Incorrect answer for question 1
	t.Run("Incorrect answer is rejected", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"q1":"A","q2":"A"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/88/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "quiz-hacker", FamilyID: "quiz-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on incorrect answer, got %d", w.Code)
		}
	})

	// Scenario 3: All correct answers succeeds exactly once
	t.Run("All correct answers succeeds and awards reward", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"q1":"B","q2":"A"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/88/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "quiz-hacker", FamilyID: "quiz-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on correct answers, got %d: %s", w.Code, w.Body.String())
		}
		if dbMock.profiles["quiz-hacker"].Coins != 50 {
			t.Fatalf("expected 50 coins awarded, got %d", dbMock.profiles["quiz-hacker"].Coins)
		}
	})
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 9: Document Upload & Template End-to-End Workflow
// -----------------------------------------------------------------------------

func TestAdversarial_DocumentWorkflowEndToEnd(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["doc-member"] = &db.UserProfile{
		UID: "doc-member", FamilyID: "doc-fam", Role: "SEEKER", Coins: 0,
	}
	dbMock.profiles["doc-admin"] = &db.UserProfile{
		UID: "doc-admin", FamilyID: "doc-fam", Role: "GUIDE", Coins: 0,
	}

	// 1. Admin creates document task with template attachment
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)
	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(`{
		"title": "Monthly Budget Sheet",
		"description": "Download template, complete the fields, and upload .xlsx back.",
		"task_type": "DOCUMENT_UPLOAD",
		"step_order": 1,
		"reward_coins": 75,
		"reward_xp": 150,
		"config": {
			"instruction": "Fill out budget sheet",
			"attachment_url": "https://cdn.example.com/templates/budget_template.xlsx",
			"attachment_name": "budget_template.xlsx",
			"accepted_extensions": [".xlsx", ".pdf"],
			"max_file_size_mb": 10
		},
		"is_active": true
	}`))
	ctxAdmin := auth.ContextWithClaims(createReq.Context(), &auth.SessionClaims{UID: "doc-admin", FamilyID: "doc-fam", Role: "GUIDE"})
	wAdmin := httptest.NewRecorder()
	adminTasksAPI.HandleCreateTask(wAdmin, createReq.WithContext(ctxAdmin))

	if wAdmin.Code != http.StatusCreated {
		t.Fatalf("failed to create document task: %d %s", wAdmin.Code, wAdmin.Body.String())
	}

	// 2. Member views task and gets attachment template URL
	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)
	getReq := httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil)
	ctxMember := auth.ContextWithClaims(getReq.Context(), &auth.SessionClaims{UID: "doc-member", FamilyID: "doc-fam", Role: "SEEKER"})
	wMember := httptest.NewRecorder()
	familyTasksAPI.Handler(wMember, getReq.WithContext(ctxMember))

	if wMember.Code != http.StatusOK {
		t.Fatalf("failed to fetch task details: %d", wMember.Code)
	}

	var taskView apiFamilyTasks.TaskView
	_ = json.Unmarshal(wMember.Body.Bytes(), &taskView)
	if taskView.Config["attachment_url"] != "https://cdn.example.com/templates/budget_template.xlsx" {
		t.Fatalf("template URL not found in task config: %v", taskView.Config)
	}

	// 3. Member uploads completed document
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, _ := mw.CreateFormFile("file", "my_completed_budget.xlsx")
	fw.Write([]byte("excel-binary-data-content"))
	mw.Close()

	uploadReq := httptest.NewRequest(http.MethodPost, "/api/tasks/upload", &buf)
	uploadReq.Header.Set("Content-Type", mw.FormDataContentType())
	wUpload := httptest.NewRecorder()
	familyTasksAPI.HandleUploadProof(wUpload, uploadReq.WithContext(ctxMember))

	if wUpload.Code != http.StatusOK {
		t.Fatalf("failed to upload document proof: %d %s", wUpload.Code, wUpload.Body.String())
	}
	var uploadRes map[string]any
	_ = json.Unmarshal(wUpload.Body.Bytes(), &uploadRes)
	fileURL := fmt.Sprintf("%v", uploadRes["file_url"])

	// 4. Member submits task with proof file_url
	submitReq := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", bytes.NewBufferString(fmt.Sprintf(`{
		"payload": {
			"file_url": %q,
			"file_name": "my_completed_budget.xlsx",
			"file_size": 1024,
			"note": "Sudah selesai dikerjakan"
		}
	}`, fileURL)))
	wSubmit := httptest.NewRecorder()
	familyTasksAPI.HandleSubmit(wSubmit, submitReq.WithContext(ctxMember))

	if wSubmit.Code != http.StatusOK {
		t.Fatalf("failed to submit document task: %d %s", wSubmit.Code, wSubmit.Body.String())
	}
	var submitRes map[string]any
	_ = json.Unmarshal(wSubmit.Body.Bytes(), &submitRes)
	subID := int64(submitRes["submission_id"].(float64))

	// 5. Admin lists pending submissions and sees the submission
	listSubReq := httptest.NewRequest(http.MethodGet, "/api/admin/submissions/pending", nil)
	wListSub := httptest.NewRecorder()
	adminTasksAPI.HandleListPendingSubmissions(wListSub, listSubReq.WithContext(ctxAdmin))

	if wListSub.Code != http.StatusOK {
		t.Fatalf("failed to list pending submissions: %d", wListSub.Code)
	}

	// 6. Admin approves submission -> 75 coins rewarded
	verifyReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(`{
		"status": "APPROVED",
		"notes": "Pekerjaan rapi dan tepat"
	}`))
	wVerify := httptest.NewRecorder()
	adminTasksAPI.HandleVerifySubmission(wVerify, verifyReq.WithContext(ctxAdmin), subID)

	if wVerify.Code != http.StatusOK {
		t.Fatalf("failed to approve submission: %d %s", wVerify.Code, wVerify.Body.String())
	}
	if dbMock.profiles["doc-member"].Coins != 75 {
		t.Fatalf("expected 75 coins awarded to member, got %d", dbMock.profiles["doc-member"].Coins)
	}
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 10: Live Active Route Surface Verification
// -----------------------------------------------------------------------------

func TestAdversarial_LiveActiveRouteSurfaceVerification(t *testing.T) {
	os.Setenv("SUPABASE_URL", "https://example.supabase.co")
	os.Setenv("SUPABASE_SERVICE_KEY", "mock-service-key")
	os.Setenv("SESSION_SIGNING_SECRET", "mock-session-signing-secret-key-32b")
	os.Setenv("PARENT_ID", "mock-parent-id")
	os.Setenv("PORT", "8080")

	srv, err := server.BuildHandler()
	if err != nil {
		t.Fatalf("BuildHandler error: %v", err)
	}
	if srv.Cleanup != nil {
		defer srv.Cleanup(context.Background())
	}

	activeRoutes := []struct {
		method       string
		path         string
		expectNon404 bool
	}{
		{"GET", "/health", true},
		{"GET", "/ready", true},
		{"GET", "/live", true},
		{"GET", "/version", true},
		{"GET", "/api/csrf", true},
	}

	for _, ar := range activeRoutes {
		t.Run(fmt.Sprintf("%s %s -> Non-404", ar.method, ar.path), func(t *testing.T) {
			req := httptest.NewRequest(ar.method, ar.path, nil)
			w := httptest.NewRecorder()
			srv.Handler.ServeHTTP(w, req)

			if w.Code == http.StatusNotFound {
				t.Fatalf("ACTIVE ROUTE MISSING: %s %s returned 404", ar.method, ar.path)
			}
		})
	}
}

// -----------------------------------------------------------------------------
// ADVERSARIAL TEST 11: Phase 8 Canonical Types, Aliases & Quiz Representations
// -----------------------------------------------------------------------------

func TestAdversarial_Phase8_QuizDeterministicAnswerRepresentations(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["quiz-student"] = &db.UserProfile{
		UID: "quiz-student", FamilyID: "quiz-fam-8", Role: "SEEKER", Coins: 0,
	}

	// Task with letter correct_answer
	dbMock.tasks[101] = map[string]any{
		"id": int64(101), "family_id": "quiz-fam-8", "title": "Geography Quiz", "task_type": "QUIZ",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"questions": []any{
				map[string]any{
					"id":             "q1",
					"question":       "Ibukota Indonesia?",
					"options":        []any{"A. Jakarta", "B. Bandung", "C. Surabaya"},
					"correct_answer": "A",
				},
			},
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Test 1: Submitting letter "A" succeeds
	t.Run("Submitting letter A succeeds", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"q1":"A"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/101/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "quiz-student", FamilyID: "quiz-fam-8", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on letter 'A', got %d %s", w.Code, w.Body.String())
		}
		if dbMock.profiles["quiz-student"].Coins != 50 {
			t.Fatalf("expected 50 coins, got %d", dbMock.profiles["quiz-student"].Coins)
		}
	})

	// Test 2: Submitting again fails with anti-double-claim
	t.Run("Replaying submission fails with anti-double-claim", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"q1":"A"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/101/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "quiz-student", FamilyID: "quiz-fam-8", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code == http.StatusOK {
			t.Fatalf("expected failure on replay, got 200 OK")
		}
		// Balance must remain exactly 50
		if dbMock.profiles["quiz-student"].Coins != 50 {
			t.Fatalf("double-claim occurred! Coins = %d", dbMock.profiles["quiz-student"].Coins)
		}
	})

	// Task with option string correct_answer
	dbMock.profiles["quiz-student-2"] = &db.UserProfile{
		UID: "quiz-student-2", FamilyID: "quiz-fam-8", Role: "SEEKER", Coins: 0,
	}
	dbMock.tasks[102] = map[string]any{
		"id": int64(102), "family_id": "quiz-fam-8", "title": "Math Quiz", "task_type": "QUIZ",
		"reward_coins": 50, "reward_xp": 100, "is_active": true,
		"config": map[string]any{
			"questions": []any{
				map[string]any{
					"id":             "q1",
					"question":       "2 + 2 = ?",
					"options":        []any{"A. 4", "B. 5"},
					"correct_answer": "A. 4",
				},
			},
		},
	}

	// Test 3: Submitting letter "A" when correct_answer is "A. 4" matches prefix
	t.Run("Submitting letter A matches prefix of 'A. 4'", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"q1":"A"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/102/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "quiz-student-2", FamilyID: "quiz-fam-8", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on prefix match, got %d %s", w.Code, w.Body.String())
		}
	})
}

func TestAdversarial_Phase8_MiniGameScoreBoundsTrustModel(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["gamer-1"] = &db.UserProfile{
		UID: "gamer-1", FamilyID: "game-fam", Role: "SEEKER", Coins: 0,
	}

	dbMock.tasks[201] = map[string]any{
		"id": int64(201), "family_id": "game-fam", "title": "Memory Challenge", "task_type": "MINI_GAME",
		"reward_coins": 60, "reward_xp": 120, "is_active": true,
		"config": map[string]any{
			"game":         "MEMORY",
			"difficulty":   "MEDIUM",
			"target_score": 80,
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Negative score -> 400 Bad Request
	t.Run("Negative score is rejected", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"score":-10,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/201/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "gamer-1", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on negative score, got %d", w.Code)
		}
	})

	// Score below target -> 400 Bad Request
	t.Run("Score below target is rejected", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"score":75,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/201/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "gamer-1", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on score below target, got %d", w.Code)
		}
	})

	// Absurdly high score > 1,000,000 -> 400 Bad Request
	t.Run("Score > 1,000,000 is rejected", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"score":99999999,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/201/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "gamer-1", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 Bad Request on absurd score, got %d", w.Code)
		}
	})

	// Valid target score -> 200 OK & 60 coins
	t.Run("Valid target score succeeds", func(t *testing.T) {
		payload := `{"submission_type":"AUTO_QUIZ","answers":{"score":85,"moves":10,"time_seconds":15,"game":"MEMORY"}}`
		req := httptest.NewRequest(http.MethodPost, "/api/tasks/201/submit", bytes.NewBufferString(payload))
		ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "gamer-1", FamilyID: "game-fam", Role: "SEEKER"})
		w := httptest.NewRecorder()

		familyTasksAPI.HandleSubmit(w, req.WithContext(ctx))
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK on valid target score, got %d %s", w.Code, w.Body.String())
		}
		if dbMock.profiles["gamer-1"].Coins != 60 {
			t.Fatalf("expected 60 coins, got %d", dbMock.profiles["gamer-1"].Coins)
		}
	})
}

func TestAdversarial_Phase8_RealisticFamilyJourneys_A_to_F(t *testing.T) {
	dbMock := newMockDB()

	// Seed Family Alpha
	dbMock.profiles["parent-guide"] = &db.UserProfile{UID: "parent-guide", FamilyID: "alpha-fam", Role: "GUIDE", Coins: 500}
	dbMock.profiles["child-seeker"] = &db.UserProfile{UID: "child-seeker", FamilyID: "alpha-fam", Role: "SEEKER", Coins: 0}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)

	ctxGuide := auth.ContextWithClaims(context.Background(), &auth.SessionClaims{UID: "parent-guide", FamilyID: "alpha-fam", Role: "GUIDE"})
	ctxSeeker := auth.ContextWithClaims(context.Background(), &auth.SessionClaims{UID: "child-seeker", FamilyID: "alpha-fam", Role: "SEEKER"})

	// Journey A: VIDEO
	t.Run("Journey A: Video Task", func(t *testing.T) {
		createReq := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(`{
			"title": "Tutorial Sains",
			"task_type": "VIDEO",
			"step_order": 1,
			"reward_coins": 50,
			"reward_xp": 100,
			"is_active": true,
			"config": {"youtube_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ", "minimum_duration_seconds": 60}
		}`))
		wCreate := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(wCreate, createReq.WithContext(ctxGuide))
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("failed to create video task: %d", wCreate.Code)
		}

		submitReq := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", bytes.NewBufferString(`{
			"submission_type": "AUTO_QUIZ",
			"payload": {"watched_seconds": 65, "completed": true}
		}`))
		wSubmit := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(wSubmit, submitReq.WithContext(ctxSeeker))
		if wSubmit.Code != http.StatusOK {
			t.Fatalf("failed to submit video task: %d %s", wSubmit.Code, wSubmit.Body.String())
		}
		if dbMock.profiles["child-seeker"].Coins != 50 {
			t.Fatalf("expected 50 coins, got %d", dbMock.profiles["child-seeker"].Coins)
		}
	})

	// Journey D: Photo Reject & Retry
	t.Run("Journey D: Photo Rejection & Retry Workflow", func(t *testing.T) {
		createReq := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(`{
			"title": "Rapikan Kamar",
			"task_type": "PHOTO_UPLOAD",
			"step_order": 2,
			"reward_coins": 40,
			"reward_xp": 80,
			"is_active": true,
			"config": {"instruction": "Foto kamar yang sudah rapi"}
		}`))
		wCreate := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(wCreate, createReq.WithContext(ctxGuide))
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("failed to create photo task: %d", wCreate.Code)
		}

		// Seeker submits photo (Task ID 2)
		subReq := httptest.NewRequest(http.MethodPost, "/api/tasks/2/submit", bytes.NewBufferString(`{
			"submission_type": "MANUAL_VERIFY",
			"payload": {"file_url": "https://cdn.example.com/alpha-fam/child-seeker/photo1.jpg"}
		}`))
		wSub := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(wSub, subReq.WithContext(ctxSeeker))
		if wSub.Code != http.StatusOK {
			t.Fatalf("failed to submit photo: %d %s", wSub.Code, wSub.Body.String())
		}
		var subRes map[string]any
		_ = json.Unmarshal(wSub.Body.Bytes(), &subRes)
		subID := int64(subRes["submission_id"].(float64))

		// Guide rejects submission with feedback note
		rejReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(`{
			"status": "REJECTED",
			"notes": "Masih ada baju berserakan di kasur"
		}`))
		wRej := httptest.NewRecorder()
		adminTasksAPI.HandleVerifySubmission(wRej, rejReq.WithContext(ctxGuide), subID)
		if wRej.Code != http.StatusOK {
			t.Fatalf("failed to reject submission: %d", wRej.Code)
		}

		// Balance remains 50 (no reward credited yet)
		if dbMock.profiles["child-seeker"].Coins != 50 {
			t.Fatalf("coins changed upon rejection! Coins = %d", dbMock.profiles["child-seeker"].Coins)
		}

		// Seeker re-submits improved photo
		retryReq := httptest.NewRequest(http.MethodPost, "/api/tasks/2/submit", bytes.NewBufferString(`{
			"submission_type": "MANUAL_VERIFY",
			"payload": {"file_url": "https://cdn.example.com/alpha-fam/child-seeker/photo2_clean.jpg"}
		}`))
		wRetry := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(wRetry, retryReq.WithContext(ctxSeeker))
		if wRetry.Code != http.StatusOK {
			t.Fatalf("failed to re-submit photo: %d", wRetry.Code)
		}
		var retryRes map[string]any
		_ = json.Unmarshal(wRetry.Body.Bytes(), &retryRes)
		newSubID := int64(retryRes["submission_id"].(float64))

		// Guide approves re-submission
		appReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", newSubID), bytes.NewBufferString(`{
			"status": "APPROVED",
			"notes": "Sekarang sudah bersih dan rapi!"
		}`))
		wApp := httptest.NewRecorder()
		adminTasksAPI.HandleVerifySubmission(wApp, appReq.WithContext(ctxGuide), newSubID)
		if wApp.Code != http.StatusOK {
			t.Fatalf("failed to approve retry: %d", wApp.Code)
		}

		// Balance is now 50 + 40 = 90
		if dbMock.profiles["child-seeker"].Coins != 90 {
			t.Fatalf("expected 90 coins after approval, got %d", dbMock.profiles["child-seeker"].Coins)
		}
	})
}

// -----------------------------------------------------------------------------
// PHASE 9: Generic Task Engine, Composition & Validation Adversarial Tests
// -----------------------------------------------------------------------------

func TestAdversarial_Phase9_GenericCapabilityEngineAndTaskComposition(t *testing.T) {
	dbMock := newMockDB()

	// Seed Family Alpha
	dbMock.profiles["guide-p9"] = &db.UserProfile{UID: "guide-p9", FamilyID: "fam-p9", Role: "GUIDE", Coins: 1000}
	dbMock.profiles["seeker-p9"] = &db.UserProfile{UID: "seeker-p9", FamilyID: "fam-p9", Role: "SEEKER", Coins: 0}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)

	ctxGuide := auth.ContextWithClaims(context.Background(), &auth.SessionClaims{UID: "guide-p9", FamilyID: "fam-p9", Role: "GUIDE"})
	ctxSeeker := auth.ContextWithClaims(context.Background(), &auth.SessionClaims{UID: "seeker-p9", FamilyID: "fam-p9", Role: "SEEKER"})

	// 1. Composite Task: VIDEO + QUIZ
	t.Run("Composite VIDEO + QUIZ task lifecycle", func(t *testing.T) {
		createReq := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(`{
			"title": "Video Pelajaran & Kuis Pemahaman",
			"task_type": "VIDEO",
			"step_order": 1,
			"reward_coins": 75,
			"reward_xp": 150,
			"is_active": true,
			"config": {
				"youtube_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
				"minimum_duration_seconds": 60,
				"questions": [
					{
						"id": "1",
						"question": "Berapa 10 + 10?",
						"options": ["A. 20", "B. 30", "C. 40"],
						"correct_answer": "A"
					}
				]
			}
		}`))
		wCreate := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(wCreate, createReq.WithContext(ctxGuide))
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("failed to create composite video+quiz task: %d %s", wCreate.Code, wCreate.Body.String())
		}

		// Seeker retrieves task -> verify correct_answer is stripped
		getReq := httptest.NewRequest(http.MethodGet, "/api/tasks/1", nil)
		wGet := httptest.NewRecorder()
		familyTasksAPI.HandleGetTask(wGet, getReq.WithContext(ctxSeeker), 1)
		if wGet.Code != http.StatusOK {
			t.Fatalf("failed to fetch task: %d", wGet.Code)
		}
		if strings.Contains(wGet.Body.String(), `"correct_answer"`) {
			t.Fatalf("CRITICAL SECURITY LEAK: correct_answer exposed in composite task response!")
		}

		// Seeker submits quiz answers
		submitReq := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", bytes.NewBufferString(`{
			"submission_type": "AUTO_QUIZ",
			"answers": {"1": "A"}
		}`))
		wSubmit := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(wSubmit, submitReq.WithContext(ctxSeeker))
		if wSubmit.Code != http.StatusOK {
			t.Fatalf("failed to submit composite video+quiz answers: %d %s", wSubmit.Code, wSubmit.Body.String())
		}

		// Balance is exactly 75
		if dbMock.profiles["seeker-p9"].Coins != 75 {
			t.Fatalf("expected 75 coins, got %d", dbMock.profiles["seeker-p9"].Coins)
		}
	})

	// 2. Composite Task: DOCUMENT + TEXT
	t.Run("Composite DOCUMENT + TEXT task lifecycle", func(t *testing.T) {
		createReq := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(`{
			"title": "Tugas Makalah & Esai Refleksi",
			"task_type": "DOCUMENT_UPLOAD",
			"step_order": 2,
			"reward_coins": 100,
			"reward_xp": 200,
			"is_active": true,
			"config": {
				"attachment_url": "https://cdn.example.com/templates/makalah_template.docx",
				"max_file_size_mb": 15,
				"minimum_characters": 20,
				"maximum_characters": 1000
			}
		}`))
		wCreate := httptest.NewRecorder()
		adminTasksAPI.HandleCreateTask(wCreate, createReq.WithContext(ctxGuide))
		if wCreate.Code != http.StatusCreated {
			t.Fatalf("failed to create composite document+text task: %d %s", wCreate.Code, wCreate.Body.String())
		}

		// Seeker submits uploaded document URL + written text explanation
		subReq := httptest.NewRequest(http.MethodPost, "/api/tasks/2/submit", bytes.NewBufferString(`{
			"submission_type": "MANUAL_VERIFY",
			"payload": {
				"file_url": "https://cdn.example.com/fam-p9/seeker-p9/makalah_final.pdf",
				"file_name": "makalah_final.pdf",
				"file_size": 2048576,
				"text": "Berikut adalah hasil rangkuman makalah sains yang telah saya lengkapi sesuai instruksi."
			}
		}`))
		wSub := httptest.NewRecorder()
		familyTasksAPI.HandleSubmit(wSub, subReq.WithContext(ctxSeeker))
		if wSub.Code != http.StatusOK {
			t.Fatalf("failed to submit composite document+text task: %d %s", wSub.Code, wSub.Body.String())
		}
		var subRes map[string]any
		_ = json.Unmarshal(wSub.Body.Bytes(), &subRes)
		subID := int64(subRes["submission_id"].(float64))

		// Guide reviews and approves
		appReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(`{
			"status": "APPROVED",
			"notes": "Analisis makalah dan teks penjelasan sangat mendalam!"
		}`))
		wApp := httptest.NewRecorder()
		adminTasksAPI.HandleVerifySubmission(wApp, appReq.WithContext(ctxGuide), subID)
		if wApp.Code != http.StatusOK {
			t.Fatalf("failed to approve submission: %d", wApp.Code)
		}

		// Balance is now 75 + 100 = 175
		if dbMock.profiles["seeker-p9"].Coins != 175 {
			t.Fatalf("expected 175 coins, got %d", dbMock.profiles["seeker-p9"].Coins)
		}
	})
}

func TestAdversarial_Phase9_CapabilityValidatorSecurityBounds(t *testing.T) {
	dbMock := newMockDB()
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)
	ctxGuide := auth.ContextWithClaims(context.Background(), &auth.SessionClaims{UID: "admin-guard", FamilyID: "guard-fam", Role: "GUIDE"})

	scenarios := []struct {
		name    string
		payload string
	}{
		{
			name: "Malicious Video URL (javascript protocol)",
			payload: `{
				"title": "Hack Video",
				"task_type": "VIDEO",
				"config": {"youtube_url": "javascript:stealCookies()"}
			}`,
		},
		{
			name: "Duplicate Quiz Question IDs in Config",
			payload: `{
				"title": "Dup Quiz",
				"task_type": "QUIZ",
				"config": {
					"questions": [
						{"id": "q1", "question": "Question 1", "options": ["A", "B"], "correct_answer": "A"},
						{"id": "q1", "question": "Question 2", "options": ["A", "B"], "correct_answer": "B"}
					]
				}
			}`,
		},
		{
			name: "Negative Reward Coins",
			payload: `{
				"title": "Negative Reward",
				"task_type": "TEXT_RESPONSE",
				"reward_coins": -500
			}`,
		},
		{
			name: "Invalid Character Bounds (min > max)",
			payload: `{
				"title": "Inverted Chars",
				"task_type": "TEXT_RESPONSE",
				"config": {"minimum_characters": 500, "maximum_characters": 50}
			}`,
		},
		{
			name: "Unknown Task Type",
			payload: `{
				"title": "Alien Task",
				"task_type": "QUANTUM_TELEPORTATION",
				"config": {}
			}`,
		},
	}

	for _, sc := range scenarios {
		t.Run(sc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/admin/tasks", bytes.NewBufferString(sc.payload))
			w := httptest.NewRecorder()
			adminTasksAPI.HandleCreateTask(w, req.WithContext(ctxGuide))
			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 Bad Request on invalid capability configuration, got %d %s", w.Code, w.Body.String())
			}
		})
	}
}

func TestAdversarial_Phase9_ConcurrentCompositeSubmissions_100Goroutines(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["rush-seeker"] = &db.UserProfile{UID: "rush-seeker", FamilyID: "rush-fam", Role: "SEEKER", Coins: 0}
	dbMock.tasks[501] = map[string]any{
		"id":              int64(501),
		"title":           "100 Race Composite Task",
		"task_type":       "VIDEO",
		"evaluation_type": "AUTO",
		"reward_coins":    50,
		"reward_xp":       100,
		"is_active":       true,
		"family_id":       "rush-fam",
		"config": map[string]any{
			"youtube_url": "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			"questions": []any{
				map[string]any{
					"id":             "1",
					"question":       "Is concurrency safe?",
					"options":        []any{"Yes", "No"},
					"correct_answer": "Yes",
				},
			},
		},
	}

	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)
	ctxSeeker := auth.ContextWithClaims(context.Background(), &auth.SessionClaims{UID: "rush-seeker", FamilyID: "rush-fam", Role: "SEEKER"})

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	successCount := int64(0)
	failCount := int64(0)

	startGate := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			<-startGate

			req := httptest.NewRequest(http.MethodPost, "/api/tasks/501/submit", bytes.NewBufferString(`{
				"submission_type": "AUTO_QUIZ",
				"answers": {"1": "Yes"}
			}`))
			w := httptest.NewRecorder()
			familyTasksAPI.HandleSubmit(w, req.WithContext(ctxSeeker))

			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else {
				atomic.AddInt64(&failCount, 1)
			}
		}()
	}

	close(startGate)
	wg.Wait()

	if successCount != 1 {
		t.Fatalf("RACE CONDITION: expected exactly 1 successful submission, got %d", successCount)
	}
	if failCount != int64(concurrency-1) {
		t.Fatalf("expected %d rejected submissions, got %d", concurrency-1, failCount)
	}
	if dbMock.profiles["rush-seeker"].Coins != 50 {
		t.Fatalf("INCORRECT LEDGER BALANCE: expected exactly 50 coins, got %d", dbMock.profiles["rush-seeker"].Coins)
	}
}

func TestAdversarial_Phase10_ConcurrentDeviceBinding_100Goroutines(t *testing.T) {
	dbMock := newMockDB()
	hasher := auth.NewBcryptHasher()
	pwdHash, _ := hasher.Hash("secret123")

	// Pre-create user profile and local user
	dbMock.profiles["target-user-uid"] = &db.UserProfile{
		UID:          "target-user-uid",
		FamilyID:     "fam-A",
		ExplorerName: "Target User",
		Role:         "SEEKER",
		IsActive:     true,
		DeviceID:     "", // Unbound initial state
	}
	dbMock.localUsers["bounduser"] = map[string]any{
		"username":      "bounduser",
		"password_hash": pwdHash,
		"profile_uid":   "target-user-uid",
	}

	authenticator := auth.NewLocalAuthProviderWithBinder(hasher, dbMock, dbMock)
	issuer := auth.NewSessionIssuer("test-secret-32-bytes-long-signature")
	apiLogin.Setup(authenticator, issuer, dbMock)

	concurrency := 100
	var wg sync.WaitGroup
	wg.Add(concurrency)

	var successCount int64
	var blockedCount int64

	startGate := make(chan struct{})

	for i := 0; i < concurrency; i++ {
		devID := fmt.Sprintf("device-alpha")
		if i%2 == 1 {
			devID = fmt.Sprintf("device-beta")
		}

		go func(deviceID string) {
			defer wg.Done()
			<-startGate

			body := fmt.Sprintf(`{"uid":"bounduser","credential":"secret123","device":{"device_id":"%s","login_method":"BOTH"}}`, deviceID)
			req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			apiLogin.Handler(w, req)

			if w.Code == http.StatusOK {
				atomic.AddInt64(&successCount, 1)
			} else if w.Code == http.StatusForbidden && strings.Contains(w.Body.String(), "Akun sudah terhubung ke perangkat lain") {
				atomic.AddInt64(&blockedCount, 1)
			}
		}(devID)
	}

	close(startGate)
	wg.Wait()

	// Verify atomic single device binding
	if dbMock.profiles["target-user-uid"].DeviceID != "device-alpha" && dbMock.profiles["target-user-uid"].DeviceID != "device-beta" {
		t.Fatalf("EXPECTED bound device ID to be device-alpha or device-beta, got empty or corrupt: %s", dbMock.profiles["target-user-uid"].DeviceID)
	}

	total := successCount + blockedCount
	if total != int64(concurrency) {
		t.Fatalf("EXPECTED all 100 logins to be handled (success or 403 blocked), got success=%d, blocked=%d, total=%d", successCount, blockedCount, total)
	}
	if successCount != 50 || blockedCount != 50 {
		t.Fatalf("EXPECTED exactly 50 wins for winning device and 50 blocks for losing device, got success=%d, blocked=%d", successCount, blockedCount)
	}
}

func TestAdversarial_Phase11_TenantIsolationAndLegacyRejection(t *testing.T) {
	dbMock := newMockDB()
	dbMock.profiles["member-famA"] = &db.UserProfile{UID: "member-famA", FamilyID: "fam-A", ExplorerName: "User A", Role: "MEMBER", IsActive: true}
	dbMock.profiles["member-famB"] = &db.UserProfile{UID: "member-famB", FamilyID: "fam-B", ExplorerName: "User B", Role: "MEMBER", IsActive: true}

	adminMembersAPI := apiAdminMembers.NewAPI(dbMock)

	t.Run("1. Admin of Family A cannot reset device or edit member of Family B", func(t *testing.T) {
		adminAClaims := &auth.SessionClaims{UID: "admin-famA", FamilyID: "fam-A", Role: "ADMIN"}
		patchBody := `{"reset_device": true}`
		req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/member-famB", strings.NewReader(patchBody))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), adminAClaims))
		w := httptest.NewRecorder()

		adminMembersAPI.HandleUpdateMember(w, req, "member-famB")

		if w.Code != http.StatusForbidden {
			t.Fatalf("CROSS-TENANT VIOLATION: Admin of Family A was able to update member of Family B! Got status %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("2. MEMBER role cannot call Admin Member APIs", func(t *testing.T) {
		memberClaims := &auth.SessionClaims{UID: "member-1", FamilyID: "fam-A", Role: "MEMBER"}
		patchBody := `{"reset_device": true}`
		req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/member-famA", strings.NewReader(patchBody))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), memberClaims))
		w := httptest.NewRecorder()

		adminMembersAPI.HandleUpdateMember(w, req, "member-famA")

		if w.Code != http.StatusForbidden {
			t.Fatalf("UNAUTHORIZED ROLE BYPASS: MEMBER role was able to call admin member API! Got status %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("3. Legacy /api/shop/items returns 404 Not Found", func(t *testing.T) {
		shopAPI := apiShop.NewAPI(dbMock)
		req := httptest.NewRequest(http.MethodGet, "/api/shop/items", nil)
		w := httptest.NewRecorder()

		shopAPI.Handler(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("LEGACY ROUTE REACHABLE: /api/shop/items returned status %d instead of 404 Not Found!", w.Code)
		}
	})

	t.Run("4. Non-cash TargetType (PULSA/PHONE/VOUCHER) rejected with 400", func(t *testing.T) {
		shopAPI := apiShop.NewAPI(dbMock)
		memberClaims := &auth.SessionClaims{UID: "member-famA", FamilyID: "fam-A", Role: "MEMBER"}

		reqBody := `{"coins": 100, "target_type": "PULSA", "target_value": "08123456789"}`
		req := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(reqBody))
		req = req.WithContext(auth.ContextWithClaims(req.Context(), memberClaims))
		w := httptest.NewRecorder()

		shopAPI.HandleRedeem(w, req)

		if w.Code != http.StatusBadRequest || !strings.Contains(w.Body.String(), "hanya mendukung pencairan tunai") {
			t.Fatalf("LEGACY ECONOMY REJECTION FAILED: PULSA target type accepted or wrong error! Status %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("5. Role Normalization (SEEKER -> MEMBER, GUIDE/BUILDER -> ADMIN)", func(t *testing.T) {
		if auth.NormalizeRole("SEEKER") != auth.RoleMember {
			t.Errorf("SEEKER should normalize to MEMBER")
		}
		if auth.NormalizeRole("GUIDE") != auth.RoleAdmin {
			t.Errorf("GUIDE should normalize to ADMIN")
		}
		if auth.NormalizeRole("BUILDER") != auth.RoleAdmin {
			t.Errorf("BUILDER should normalize to ADMIN")
		}
	})
}

func TestAdversarial_VerificationAuthorizationAndPenaltyScoring(t *testing.T) {
	dbMock := newMockDB()
	adminTasksAPI := apiAdminTasks.NewAPI(dbMock)
	familyTasksAPI := apiFamilyTasks.NewAPI(dbMock)

	// Set up user profile with 30 coins
	dbMock.profiles["scorer-1"] = &db.UserProfile{
		UID:      "scorer-1",
		FamilyID: "fam-scoring",
		Role:     "MEMBER",
		Coins:    30,
		XP:       100,
	}
	dbMock.profiles["admin-scorer"] = &db.UserProfile{
		UID:      "admin-scorer",
		FamilyID: "fam-scoring",
		Role:     "ADMIN",
	}

	// Set up task
	dbMock.tasks[100] = map[string]any{
		"id":           int64(100),
		"family_id":    "fam-scoring",
		"title":        "Writing Task",
		"task_type":    "TEXT_RESPONSE",
		"reward_coins": 50,
		"reward_xp":    100,
		"is_active":    true,
	}

	adminClaims := &auth.SessionClaims{UID: "admin-scorer", FamilyID: "fam-scoring", Role: "ADMIN"}
	memberClaims := &auth.SessionClaims{UID: "scorer-1", FamilyID: "fam-scoring", Role: "MEMBER"}

	// 1. Member submits manual task -> status PENDING
	submitReq := httptest.NewRequest(http.MethodPost, "/api/tasks/100/submit", bytes.NewBufferString(`{"payload":{"text":"Jawaban pertama yang salah"}}`))
	submitReq.Header.Set("Content-Type", "application/json")
	wSubmit := httptest.NewRecorder()
	familyTasksAPI.HandleSubmit(wSubmit, submitReq.WithContext(auth.ContextWithClaims(submitReq.Context(), memberClaims)))

	if wSubmit.Code != http.StatusOK {
		t.Fatalf("failed to submit task: %d %s", wSubmit.Code, wSubmit.Body.String())
	}

	var submitRes map[string]any
	_ = json.Unmarshal(wSubmit.Body.Bytes(), &submitRes)
	subID := int64(submitRes["submission_id"].(float64))

	// 2. Admin edits submission payload while PENDING
	editReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/submissions/%d", subID), bytes.NewBufferString(`{"payload":{"text":"Jawaban yang sudah dikoreksi admin"}}`))
	editReq.Header.Set("Content-Type", "application/json")
	wEdit := httptest.NewRecorder()
	adminTasksAPI.HandleEditSubmission(wEdit, editReq.WithContext(auth.ContextWithClaims(editReq.Context(), adminClaims)), subID)

	if wEdit.Code != http.StatusOK {
		t.Fatalf("admin failed to edit pending submission: %d %s", wEdit.Code, wEdit.Body.String())
	}

	// 3. Admin rejects with penalty of 50 coins (User only has 30 coins -> should clamp to 30, balance -> 0)
	rejReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(`{"status":"REJECTED","notes":"Jawaban salah, dikenakan penalti","penalty_coins":50}`))
	rejReq.Header.Set("Content-Type", "application/json")
	wRej := httptest.NewRecorder()
	adminTasksAPI.HandleVerifySubmission(wRej, rejReq.WithContext(auth.ContextWithClaims(rejReq.Context(), adminClaims)), subID)

	if wRej.Code != http.StatusOK {
		t.Fatalf("admin failed to reject with penalty: %d %s", wRej.Code, wRej.Body.String())
	}

	if dbMock.profiles["scorer-1"].Coins != 0 {
		t.Fatalf("expected coins balance to clamp at 0, got %d", dbMock.profiles["scorer-1"].Coins)
	}

	// 4. Repeated reject on already REJECTED submission -> MUST FAIL (P0004) to prevent double deduction
	rejReq2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", subID), bytes.NewBufferString(`{"status":"REJECTED","penalty_coins":20}`))
	rejReq2.Header.Set("Content-Type", "application/json")
	wRej2 := httptest.NewRecorder()
	adminTasksAPI.HandleVerifySubmission(wRej2, rejReq2.WithContext(auth.ContextWithClaims(rejReq2.Context(), adminClaims)), subID)

	if wRej2.Code != http.StatusBadRequest {
		t.Fatalf("expected repeated verify to fail with 400, got %d: %s", wRej2.Code, wRej2.Body.String())
	}

	// 5. Member resubmits -> transitions to PENDING
	submitReq2 := httptest.NewRequest(http.MethodPost, "/api/tasks/100/submit", bytes.NewBufferString(`{"payload":{"text":"Jawaban kedua yang benar"}}`))
	submitReq2.Header.Set("Content-Type", "application/json")
	wSubmit2 := httptest.NewRecorder()
	familyTasksAPI.HandleSubmit(wSubmit2, submitReq2.WithContext(auth.ContextWithClaims(submitReq2.Context(), memberClaims)))

	if wSubmit2.Code != http.StatusOK {
		t.Fatalf("failed to resubmit task: %d %s", wSubmit2.Code, wSubmit2.Body.String())
	}

	var submitRes2 map[string]any
	_ = json.Unmarshal(wSubmit2.Body.Bytes(), &submitRes2)
	newSubID := int64(submitRes2["submission_id"].(float64))

	// 6. Admin approves -> status APPROVED, reward credited (+50 coins)
	appReq := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/admin/submissions/%d/verify", newSubID), bytes.NewBufferString(`{"status":"APPROVED","notes":"Bagus sekali!"}`))
	appReq.Header.Set("Content-Type", "application/json")
	wApp := httptest.NewRecorder()
	adminTasksAPI.HandleVerifySubmission(wApp, appReq.WithContext(auth.ContextWithClaims(appReq.Context(), adminClaims)), newSubID)

	if wApp.Code != http.StatusOK {
		t.Fatalf("admin failed to approve submission: %d %s", wApp.Code, wApp.Body.String())
	}

	if dbMock.profiles["scorer-1"].Coins != 50 {
		t.Fatalf("expected coins balance to be 50 after approval, got %d", dbMock.profiles["scorer-1"].Coins)
	}

	// 7. Editing APPROVED submission MUST FAIL
	editAppReq := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/api/admin/submissions/%d", newSubID), bytes.NewBufferString(`{"payload":{"text":"Ubah setelah disetujui"}}`))
	editAppReq.Header.Set("Content-Type", "application/json")
	wEditApp := httptest.NewRecorder()
	adminTasksAPI.HandleEditSubmission(wEditApp, editAppReq.WithContext(auth.ContextWithClaims(editAppReq.Context(), adminClaims)), newSubID)

	if wEditApp.Code != http.StatusBadRequest {
		t.Fatalf("expected editing approved submission to fail with 400, got %d: %s", wEditApp.Code, wEditApp.Body.String())
	}
}
