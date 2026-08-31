package shop

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

type API struct {
	client db.SupabaseClient
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{client: client}
}

type RewardCatalogItem struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Category    string `json:"category"`
	CostCoins   int    `json:"cost_coins"`
	IconName    string `json:"icon_name"`
	IsAvailable bool   `json:"is_available"`
}

type ClaimView struct {
	ID            int64   `json:"id"`
	UserUID       string  `json:"user_uid"`
	UserName      string  `json:"user_name,omitempty"`
	RewardID      *int64  `json:"reward_id,omitempty"`
	RewardTitle   string  `json:"reward_title,omitempty"`
	CoinsRedeemed int     `json:"coins_redeemed"`
	TargetType    string  `json:"target_type"`
	TargetValue   string  `json:"target_value"`
	Status        string  `json:"status"` // "PENDING", "APPROVED", "REJECTED"
	AdminNotes    *string `json:"admin_notes,omitempty"`
	CreatedAt     string  `json:"created_at"`
	ProcessedAt   *string `json:"processed_at,omitempty"`
}

func (a *API) HandleGetCatalog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	raw, err := a.client.Get(ctx, "odyssey_reward_catalog", "is_available=eq.true&order=cost_coins.asc")
	if err != nil {
		shared.WriteJSONError(w, "failed to get reward catalog: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (a *API) HandleRedeem(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return
	}
	uid := claims.UID

	var req struct {
		RewardID    *int64 `json:"reward_id"`
		Coins       int    `json:"coins"`
		TargetType  string `json:"target_type"`  // EWALLET, PHONE, BANK, CASH, etc.
		TargetValue string `json:"target_value"` // e.g. "08123456789 (GoPay)"
	}
	if err := shared.ReadJSON(r, &req); err != nil {
		shared.WriteJSONError(w, "invalid json payload: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Coins <= 0 {
		shared.WriteJSONError(w, "jumlah koin harus lebih besar dari 0", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.TargetType) == "" || strings.TrimSpace(req.TargetValue) == "" {
		shared.WriteJSONError(w, "tipe tujuan dan nomor/rekening penukaran wajib diisi", http.StatusBadRequest)
		return
	}

	rpcRes, err := a.client.RPC(ctx, "odyssey_create_claim", map[string]any{
		"p_user_uid":     uid,
		"p_coins":        req.Coins,
		"p_target_type":  req.TargetType,
		"p_target_value": req.TargetValue,
		"p_reward_id":    req.RewardID,
	})
	if err != nil {
		shared.WriteJSONError(w, "gagal mengajukan penukaran: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rpcRes)
}

func (a *API) HandleGetUserClaims(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, ok := auth.ClaimsFromContext(ctx)
	if !ok || claims == nil {
		shared.WriteUnauthorized(w)
		return
	}
	uid := claims.UID

	raw, err := a.client.Get(ctx, "odyssey_claims", fmt.Sprintf("user_uid=eq.%s&order=created_at.desc", uid))
	if err != nil {
		shared.WriteJSONError(w, "failed to get claims: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(raw)
}

func (a *API) HandleAdminListClaims(w http.ResponseWriter, r *http.Request) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims == nil || claims.Role != "GUIDE" {
		shared.WriteJSONError(w, "akses ditolak: hanya untuk admin", http.StatusForbidden)
		return
	}
	ctx := r.Context()

	statusFilter := r.URL.Query().Get("status")
	params := "order=created_at.desc"
	if statusFilter != "" {
		params = fmt.Sprintf("status=eq.%s&order=created_at.desc", statusFilter)
	}

	raw, err := a.client.Get(ctx, "odyssey_claims", params)
	if err != nil {
		shared.WriteJSONError(w, "failed to get claims: "+err.Error(), http.StatusInternalServerError)
		return
	}

	type RawClaim struct {
		ID            int64   `json:"id"`
		UserUID       string  `json:"user_uid"`
		RewardID      *int64  `json:"reward_id"`
		CoinsRedeemed int     `json:"coins_redeemed"`
		TargetType    string  `json:"target_type"`
		TargetValue   string  `json:"target_value"`
		Status        string  `json:"status"`
		AdminNotes    *string `json:"admin_notes"`
		CreatedAt     string  `json:"created_at"`
		ProcessedAt   *string `json:"processed_at"`
	}
	var claimList []RawClaim
	_ = json.Unmarshal(raw, &claimList)

	// Fetch user names scoped to admin's family
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

	// Filter claims by admin's family
	filteredClaims := make([]RawClaim, 0, len(claimList))
	for _, c := range claimList {
		if claims.FamilyID == "" || familyMemberUIDs[c.UserUID] {
			filteredClaims = append(filteredClaims, c)
		}
	}

	// Fetch catalog titles
	catRaw, _ := a.client.Get(ctx, "odyssey_reward_catalog", "select=id,title")
	type CatRow struct {
		ID    int64  `json:"id"`
		Title string `json:"title"`
	}
	var cats []CatRow
	_ = json.Unmarshal(catRaw, &cats)
	catMap := make(map[int64]string)
	for _, c := range cats {
		catMap[c.ID] = c.Title
	}

	views := make([]ClaimView, len(filteredClaims))
	for i, c := range filteredClaims {
		name := profMap[c.UserUID]
		if name == "" {
			name = c.UserUID
		}
		var title string
		if c.RewardID != nil {
			title = catMap[*c.RewardID]
		}
		views[i] = ClaimView{
			ID:            c.ID,
			UserUID:       c.UserUID,
			UserName:      name,
			RewardID:      c.RewardID,
			RewardTitle:   title,
			CoinsRedeemed: c.CoinsRedeemed,
			TargetType:    c.TargetType,
			TargetValue:   c.TargetValue,
			Status:        c.Status,
			AdminNotes:    c.AdminNotes,
			CreatedAt:     c.CreatedAt,
			ProcessedAt:   c.ProcessedAt,
		}
	}

	shared.WriteJSON(w, http.StatusOK, views)
}

func (a *API) HandleAdminProcessClaim(w http.ResponseWriter, r *http.Request, claimID int64) {
	claims, ok := auth.ClaimsFromContext(r.Context())
	if !ok || claims == nil || claims.Role != "GUIDE" {
		shared.WriteJSONError(w, "akses ditolak: hanya untuk admin", http.StatusForbidden)
		return
	}
	ctx := r.Context()

	// Verify claim belongs to admin's family
	cRaw, err := a.client.Get(ctx, "odyssey_claims", fmt.Sprintf("id=eq.%d", claimID))
	if err != nil || len(cRaw) == 0 {
		shared.WriteJSONError(w, "klaim tidak ditemukan", http.StatusNotFound)
		return
	}
	type ClaimCheck struct {
		UserUID string `json:"user_uid"`
	}
	var cChecks []ClaimCheck
	_ = json.Unmarshal(cRaw, &cChecks)
	if len(cChecks) > 0 && claims.FamilyID != "" {
		pRaw, _ := a.client.Get(ctx, "odyssey_user_profiles", fmt.Sprintf("uid=eq.%s", cChecks[0].UserUID))
		type ProfileCheck struct {
			FamilyID string `json:"family_id"`
		}
		var pChecks []ProfileCheck
		_ = json.Unmarshal(pRaw, &pChecks)
		if len(pChecks) > 0 && pChecks[0].FamilyID != "" && pChecks[0].FamilyID != claims.FamilyID {
			shared.WriteJSONError(w, "akses ditolak: klaim bukan milik anggota keluarga Anda", http.StatusForbidden)
			return
		}
	}

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

	rpcRes, err := a.client.RPC(ctx, "odyssey_process_claim", map[string]any{
		"p_claim_id":    claimID,
		"p_status":      req.Status,
		"p_admin_notes": req.Notes,
	})
	if err != nil {
		shared.WriteJSONError(w, "gagal memproses klaim: "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(rpcRes)
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")

	if path == "/api/shop/items" && r.Method == http.MethodGet {
		a.HandleGetCatalog(w, r)
		return
	}
	if path == "/api/shop/redeem" && r.Method == http.MethodPost {
		a.HandleRedeem(w, r)
		return
	}
	if path == "/api/shop/claims" && r.Method == http.MethodGet {
		a.HandleGetUserClaims(w, r)
		return
	}
	if path == "/api/admin/claims" && r.Method == http.MethodGet {
		a.HandleAdminListClaims(w, r)
		return
	}
	if strings.HasPrefix(path, "/api/admin/claims/") && strings.HasSuffix(path, "/process") && r.Method == http.MethodPost {
		parts := strings.Split(path, "/")
		if len(parts) >= 5 {
			claimID, err := strconv.ParseInt(parts[4], 10, 64)
			if err == nil {
				a.HandleAdminProcessClaim(w, r, claimID)
				return
			}
		}
	}

	http.NotFound(w, r)
}
