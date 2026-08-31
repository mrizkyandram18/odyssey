package admin_members

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type API struct {
	client db.SupabaseClient
	hasher auth.PasswordHasher
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{
		client: client,
		hasher: auth.NewBcryptHasher(),
	}
}

type MemberView struct {
	UID          string    `json:"uid"`
	FamilyID     string    `json:"family_id"`
	ExplorerName string    `json:"explorer_name"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	Level        int       `json:"level"`
	XP           int64     `json:"xp"`
	Coins        int64     `json:"coins"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateMemberRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ExplorerName string `json:"explorer_name"`
	Role         string `json:"role,omitempty"`
}

type UpdateMemberRequest struct {
	ExplorerName *string `json:"explorer_name,omitempty"`
	Role         *string `json:"role,omitempty"`
	IsActive     *bool   `json:"is_active,omitempty"`
	Password     *string `json:"password,omitempty"`
	ResetDevice  *bool   `json:"reset_device,omitempty"`
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

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

func generateID(prefix string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s_%d_%s", prefix, time.Now().Unix(), hex.EncodeToString(b))
}

func (a *API) HandleListMembers(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	// 1. Fetch user profiles for admin's family
	profParams := "order=created_at.desc"
	if claims.FamilyID != "" {
		profParams = fmt.Sprintf("family_id=eq.%s&order=created_at.desc", claims.FamilyID)
	}

	profRaw, err := a.client.Get(ctx, "odyssey_user_profiles", profParams)
	if err != nil {
		shared.WriteJSONError(w, "gagal mengambil daftar anggota: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type ProfRow struct {
		UID          string    `json:"uid"`
		FamilyID     string    `json:"family_id"`
		ExplorerName string    `json:"explorer_name"`
		Role         string    `json:"role"`
		IsActive     bool      `json:"is_active"`
		Level        int       `json:"level"`
		XP           int64     `json:"xp"`
		Coins        int64     `json:"coins"`
		CreatedAt    time.Time `json:"created_at"`
	}
	var profs []ProfRow
	_ = json.Unmarshal(profRaw, &profs)

	if len(profs) == 0 {
		shared.WriteJSON(w, http.StatusOK, []MemberView{})
		return
	}

	// 2. Fetch local user credentials to map usernames
	localRaw, _ := a.client.Get(ctx, "odyssey_local_users", "select=username,profile_uid")
	type LocalRow struct {
		Username   string `json:"username"`
		ProfileUID string `json:"profile_uid"`
	}
	var locals []LocalRow
	_ = json.Unmarshal(localRaw, &locals)

	userMap := make(map[string]string)
	for _, l := range locals {
		userMap[l.ProfileUID] = l.Username
	}

	res := make([]MemberView, len(profs))
	for i, p := range profs {
		username := userMap[p.UID]
		if username == "" {
			username = p.UID
		}
		res[i] = MemberView{
			UID:          p.UID,
			FamilyID:     p.FamilyID,
			ExplorerName: p.ExplorerName,
			Username:     username,
			Role:         p.Role,
			IsActive:     p.IsActive,
			Level:        p.Level,
			XP:           p.XP,
			Coins:        p.Coins,
			CreatedAt:    p.CreatedAt,
		}
	}

	shared.WriteJSON(w, http.StatusOK, res)
}

func (a *API) HandleCreateMember(w http.ResponseWriter, r *http.Request) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	var req CreateMemberRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.ExplorerName = strings.TrimSpace(req.ExplorerName)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || len(req.Username) < 3 {
		shared.WriteJSONError(w, "username minimal 3 karakter", http.StatusBadRequest)
		return
	}
	if !usernameRegex.MatchString(req.Username) {
		shared.WriteJSONError(w, "username hanya boleh berisi huruf, angka, titik, strip, dan underscore", http.StatusBadRequest)
		return
	}
	if req.Password == "" || len(req.Password) < 6 {
		shared.WriteJSONError(w, "password minimal 6 karakter", http.StatusBadRequest)
		return
	}
	if req.ExplorerName == "" {
		shared.WriteJSONError(w, "nama anggota wajib diisi", http.StatusBadRequest)
		return
	}

	role := string(auth.NormalizeRole(req.Role))

	// 1. Check username uniqueness
	existingRaw, err := a.client.Get(ctx, "odyssey_local_users", fmt.Sprintf("username=eq.%s", req.Username))
	if err == nil && len(existingRaw) > 2 && string(existingRaw) != "[]" {
		shared.WriteJSONError(w, "username sudah digunakan, pilih username lain", http.StatusBadRequest)
		return
	}

	// 2. Hash password server-side with bcrypt
	passHash, err := a.hasher.Hash(req.Password)
	if err != nil {
		shared.WriteJSONError(w, "gagal memproses enkripsi password", http.StatusInternalServerError)
		return
	}

	familyID := claims.FamilyID
	if familyID == "" {
		familyID = "family_default"
	}

	uid := generateID("usr")

	// 3. Insert user profile
	now := time.Now().UTC()
	profPayload := map[string]any{
		"uid":                  uid,
		"family_id":            familyID,
		"explorer_name":        req.ExplorerName,
		"role":                 role,
		"level":                1,
		"xp":                   0,
		"coins":                0,
		"streak_days":          0,
		"is_active":            true,
		"must_change_password": true,
		"created_at":           now.Format(time.RFC3339),
		"updated_at":           now.Format(time.RFC3339),
	}

	_, err = a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_user_profiles", profPayload, "", "return=representation")
	if err != nil {
		shared.WriteJSONError(w, "gagal membuat profil anggota: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Insert local auth user
	localID := generateID("loc")
	localPayload := map[string]any{
		"id":            localID,
		"username":      req.Username,
		"password_hash": passHash,
		"profile_uid":   uid,
		"created_at":    now.Format(time.RFC3339),
		"updated_at":    now.Format(time.RFC3339),
	}

	_, err = a.client.MutateAtomic(ctx, http.MethodPost, "odyssey_local_users", localPayload, "", "return=representation")
	if err != nil {
		// Rollback profile on auth insertion failure
		_, _ = a.client.Mutate(ctx, http.MethodDelete, "odyssey_user_profiles", nil, fmt.Sprintf("uid=eq.%s", uid))
		shared.WriteJSONError(w, "gagal membuat kredensial login: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, http.StatusCreated, MemberView{
		UID:          uid,
		FamilyID:     familyID,
		ExplorerName: req.ExplorerName,
		Username:     req.Username,
		Role:         role,
		IsActive:     true,
		Level:        1,
		XP:           0,
		Coins:        0,
		CreatedAt:    now,
	})
}

func (a *API) HandleUpdateMember(w http.ResponseWriter, r *http.Request, targetUID string) {
	claims, ok := a.requireAdmin(w, r)
	if !ok {
		return
	}
	ctx := r.Context()

	// Tenant check: target profile must belong to admin's family
	targetRaw, err := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	if err != nil || len(targetRaw) == 0 {
		shared.WriteJSONError(w, "anggota tidak ditemukan", http.StatusNotFound)
		return
	}

	type ProfileCheck struct {
		UID      string `json:"uid"`
		FamilyID string `json:"family_id"`
	}
	var checks []ProfileCheck
	_ = json.Unmarshal(targetRaw, &checks)
	if len(checks) > 0 && claims.FamilyID != "" && checks[0].FamilyID != "" && checks[0].FamilyID != claims.FamilyID {
		shared.WriteJSONError(w, "akses ditolak: anggota bukan bagian dari keluarga Anda", http.StatusForbidden)
		return
	}

	var req UpdateMemberRequest
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json body: "+err.Error(), http.StatusBadRequest)
		return
	}

	profPatch := map[string]any{}
	if req.ExplorerName != nil {
		name := strings.TrimSpace(*req.ExplorerName)
		if name == "" {
			shared.WriteJSONError(w, "nama anggota tidak boleh kosong", http.StatusBadRequest)
			return
		}
		profPatch["explorer_name"] = name
	}
	if req.Role != nil {
		role := strings.ToUpper(strings.TrimSpace(*req.Role))
		if role != "MEMBER" && role != "ADMIN" && role != "SEEKER" && role != "GUIDE" {
			shared.WriteJSONError(w, "role tidak valid", http.StatusBadRequest)
			return
		}
		profPatch["role"] = role
	}
	if req.IsActive != nil {
		profPatch["is_active"] = *req.IsActive
	}
	if req.ResetDevice != nil && *req.ResetDevice {
		profPatch["device_id"] = nil
		profPatch["device_bound_at"] = nil
		_, _ = a.client.RPC(ctx, "odyssey_admin_reset_device", map[string]any{"p_target_uid": targetUID})
	}

	// Handle password reset if provided
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		pass := strings.TrimSpace(*req.Password)
		if len(pass) < 6 {
			shared.WriteJSONError(w, "password baru minimal 6 karakter", http.StatusBadRequest)
			return
		}
		profPatch["must_change_password"] = true

		hash, err := a.hasher.Hash(pass)
		if err != nil {
			shared.WriteJSONError(w, "gagal memproses password", http.StatusInternalServerError)
			return
		}
		localPatch := map[string]any{
			"password_hash": hash,
			"updated_at":    time.Now().UTC().Format(time.RFC3339),
		}
		_, err = a.client.Mutate(ctx, http.MethodPatch, "odyssey_local_users", localPatch, fmt.Sprintf("profile_uid=eq.%s", targetUID))
		if err != nil {
			shared.WriteJSONError(w, "gagal update password anggota: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if len(profPatch) > 0 {
		profPatch["updated_at"] = time.Now().UTC().Format(time.RFC3339)
		_, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_profiles", profPatch, fmt.Sprintf("uid=eq.%s", targetUID))
		if err != nil {
			shared.WriteJSONError(w, "gagal update profil anggota: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Return reloaded member view
	updatedRaw, _ := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", targetUID))
	var updatedProfs []db.UserProfile
	_ = json.Unmarshal(updatedRaw, &updatedProfs)

	var username string
	localRaw, _ := a.client.Get(ctx, "odyssey_local_users", fmt.Sprintf("profile_uid=eq.%s", targetUID))
	var locals []struct {
		Username string `json:"username"`
	}
	_ = json.Unmarshal(localRaw, &locals)
	if len(locals) > 0 {
		username = locals[0].Username
	} else {
		username = targetUID
	}

	if len(updatedProfs) > 0 {
		up := updatedProfs[0]
		shared.WriteJSON(w, http.StatusOK, MemberView{
			UID:          up.UID,
			FamilyID:     up.FamilyID,
			ExplorerName: up.ExplorerName,
			Username:     username,
			Role:         up.Role,
			IsActive:     up.IsActive,
			Level:        up.Level,
			XP:           up.XP,
			Coins:        up.Coins,
			CreatedAt:    up.CreatedAt,
		})
		return
	}

	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	if path == "/api/admin/members" {
		if r.Method == http.MethodGet {
			a.HandleListMembers(w, r)
			return
		}
		if r.Method == http.MethodPost {
			a.HandleCreateMember(w, r)
			return
		}
	}

	if strings.HasPrefix(path, "/api/admin/members/") {
		targetUID := strings.TrimPrefix(path, "/api/admin/members/")
		if targetUID != "" && r.Method == http.MethodPatch {
			a.HandleUpdateMember(w, r, targetUID)
			return
		}
	}

	http.NotFound(w, r)
}
