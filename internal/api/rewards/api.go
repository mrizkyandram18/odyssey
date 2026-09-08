package rewards

import (
	"encoding/json"
	"net/http"
	"strings"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/shared"
)

// API serves reward-ticket + collection endpoints for the authenticated user.
// All state is uid-scoped to the session claims (family isolation by construction).
// Display copy must use layman terms (Tiket Hadiah / Buka Hadiah / Koleksi).
type API struct {
	client db.SupabaseClient
}

func NewAPI(client db.SupabaseClient) *API {
	return &API{client: client}
}

func (a *API) claims(w http.ResponseWriter, r *http.Request) (*auth.SessionClaims, bool) {
	c, ok := auth.ClaimsFromRequest(r)
	if !ok || c == nil || strings.TrimSpace(c.UID) == "" {
		shared.WriteUnauthorized(w)
		return nil, false
	}
	if c.Kind != string(auth.SessionKindUser) {
		shared.WriteForbidden(w)
		return nil, false
	}
	return c, true
}

func (a *API) Handler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimSuffix(r.URL.Path, "/")
	switch {
	case path == "/api/rewards" && r.Method == http.MethodGet:
		a.handleStatus(w, r)
	case path == "/api/rewards/claim" && r.Method == http.MethodPost:
		a.handleClaim(w, r)
	case path == "/api/rewards/open" && r.Method == http.MethodPost:
		a.handleOpen(w, r)
	case path == "/api/rewards/collection" && r.Method == http.MethodGet:
		a.handleCollection(w, r)
	case path == "/api/rewards/equip" && r.Method == http.MethodPost:
		a.handleEquip(w, r)
	default:
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
	}
}

type statusResponse struct {
	Tickets int `json:"tickets"`
}

func (a *API) handleStatus(w http.ResponseWriter, r *http.Request) {
	c, ok := a.claims(w, r)
	if !ok {
		return
	}
	raw, err := a.client.Get(r.Context(), "reward_tickets",
		"user_uid=eq."+c.UID+"&status=eq.GRANTED&select=ticket_date")
	if err != nil {
		shared.WriteJSONError(w, "gagal memuat tiket", http.StatusInternalServerError)
		return
	}
	var rows []map[string]any
	_ = json.Unmarshal(raw, &rows)
	shared.WriteJSON(w, http.StatusOK, statusResponse{Tickets: len(rows)})
}

func (a *API) handleClaim(w http.ResponseWriter, r *http.Request) {
	c, ok := a.claims(w, r)
	if !ok {
		return
	}
	raw, err := a.client.RPC(r.Context(), "odyssey_claim_daily_ticket", map[string]any{"p_user_uid": c.UID})
	if err != nil {
		shared.WriteJSONError(w, claimErrorMessage(err), http.StatusBadRequest)
		return
	}
	var res struct {
		Granted bool `json:"granted"`
		Tickets int  `json:"tickets"`
	}
	_ = json.Unmarshal(raw, &res)
	shared.WriteJSON(w, http.StatusOK, res)
}

type openResponse struct {
	CosmeticID string `json:"cosmetic_id"`
	Slot       string `json:"slot"`
	Asset      string `json:"asset"`
	Tier       int    `json:"tier"`
	IsNew      bool   `json:"is_new"`
}

func (a *API) handleOpen(w http.ResponseWriter, r *http.Request) {
	c, ok := a.claims(w, r)
	if !ok {
		return
	}
	raw, err := a.client.RPC(r.Context(), "odyssey_open_reward_box", map[string]any{"p_user_uid": c.UID})
	if err != nil {
		shared.WriteJSONError(w, openErrorMessage(err), http.StatusBadRequest)
		return
	}
	var res openResponse
	if err := json.Unmarshal(raw, &res); err != nil || res.CosmeticID == "" {
		shared.WriteJSONError(w, "gagal membuka hadiah", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, res)
}

type collectionItem struct {
	CosmeticID string `json:"cosmetic_id"`
	Slot       string `json:"slot"`
	Asset      string `json:"asset"`
	Tier       int    `json:"tier"`
	Owned      bool   `json:"owned"`
	Equipped   bool   `json:"equipped"`
}

func (a *API) handleCollection(w http.ResponseWriter, r *http.Request) {
	c, ok := a.claims(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	catRaw, err := a.client.Get(ctx, "cosmetic_items", "is_active=eq.true&order=slot.asc,id.asc&select=id,slot,asset")
	if err != nil {
		shared.WriteJSONError(w, "gagal memuat koleksi", http.StatusInternalServerError)
		return
	}
	var catalog []struct {
		ID    string `json:"id"`
		Slot  string `json:"slot"`
		Asset string `json:"asset"`
	}
	_ = json.Unmarshal(catRaw, &catalog)

	ownRaw, _ := a.client.Get(ctx, "user_cosmetics", "user_uid=eq."+c.UID+"&select=cosmetic_id,tier,equipped")
	var owned []struct {
		CosmeticID string `json:"cosmetic_id"`
		Tier       int    `json:"tier"`
		Equipped   bool   `json:"equipped"`
	}
	_ = json.Unmarshal(ownRaw, &owned)
	ownedMap := make(map[string]struct {
		Tier     int
		Equipped bool
	}, len(owned))
	for _, o := range owned {
		ownedMap[o.CosmeticID] = struct {
			Tier     int
			Equipped bool
		}{o.Tier, o.Equipped}
	}

	profRaw, _ := a.client.Get(ctx, "odyssey_user_profiles", "uid=eq."+c.UID+"&select=avatar_frame,equipped_explorer_effect")
	var profs []struct {
		AvatarFrame            string `json:"avatar_frame"`
		EquippedExplorerEffect string `json:"equipped_explorer_effect"`
	}
	_ = json.Unmarshal(profRaw, &profs)
	curFrame := "none"
	curEffect := "none"
	if len(profs) > 0 {
		if profs[0].AvatarFrame != "" {
			curFrame = profs[0].AvatarFrame
		}
		if profs[0].EquippedExplorerEffect != "" {
			curEffect = profs[0].EquippedExplorerEffect
		}
	}

	items := make([]collectionItem, 0, len(catalog))
	for _, cat := range catalog {
		o, hasOwned := ownedMap[cat.ID]
		isEquipped := false
		if hasOwned {
			if cat.Slot == "frame" && curFrame == cat.Asset {
				isEquipped = true
			} else if cat.Slot == "effect" && curEffect == cat.Asset {
				isEquipped = true
			} else if o.Equipped {
				isEquipped = true
			}
		}
		items = append(items, collectionItem{
			CosmeticID: cat.ID,
			Slot:       cat.Slot,
			Asset:      cat.Asset,
			Tier:       o.Tier,
			Owned:      hasOwned,
			Equipped:   isEquipped,
		})
	}
	shared.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (a *API) handleEquip(w http.ResponseWriter, r *http.Request) {
	c, ok := a.claims(w, r)
	if !ok {
		return
	}
	var req struct {
		CosmeticID string `json:"cosmetic_id"`
	}
	if err := shared.ReadJSON(r, &req); err != nil || strings.TrimSpace(req.CosmeticID) == "" {
		shared.WriteJSONError(w, "cosmetic_id wajib diisi", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	ownRaw, err := a.client.Get(ctx, "user_cosmetics",
		"user_uid=eq."+c.UID+"&cosmetic_id=eq."+req.CosmeticID+"&select=cosmetic_id")
	if err != nil || len(ownRaw) <= 2 || strings.TrimSpace(string(ownRaw)) == "[]" {
		shared.WriteJSONError(w, "hadiah belum dimiliki", http.StatusForbidden)
		return
	}
	catRaw, err := a.client.Get(ctx, "cosmetic_items",
		"id=eq."+req.CosmeticID+"&select=slot,asset")
	if err != nil {
		shared.WriteJSONError(w, "hadiah tidak valid", http.StatusBadRequest)
		return
	}
	var cat []struct {
		Slot  string `json:"slot"`
		Asset string `json:"asset"`
	}
	_ = json.Unmarshal(catRaw, &cat)
	if len(cat) == 0 {
		shared.WriteJSONError(w, "hadiah tidak valid", http.StatusBadRequest)
		return
	}
	patch := map[string]any{}
	switch cat[0].Slot {
	case "frame":
		patch["avatar_frame"] = cat[0].Asset
	case "effect":
		patch["equipped_explorer_effect"] = cat[0].Asset
	default:
		shared.WriteJSONError(w, "hadiah tidak valid", http.StatusBadRequest)
		return
	}
	if _, err := a.client.Mutate(ctx, http.MethodPatch, "odyssey_user_profiles", patch, "uid=eq."+c.UID); err != nil {
		shared.WriteJSONError(w, "gagal memasang hadiah: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// Mark equipped item (best-effort; profile columns are the display SOT).
	_, _ = a.client.Mutate(ctx, http.MethodPatch, "user_cosmetics", map[string]any{"equipped": true},
		"user_uid=eq."+c.UID+"&cosmetic_id=eq."+req.CosmeticID)
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "success"})
}

func claimErrorMessage(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "P0021"):
		return "akun Anda nonaktif, silakan hubungi admin"
	case strings.Contains(msg, "P0023"):
		return "Selesaikan minimal 1 tugas hari ini untuk mendapatkan Tiket Hadiah"
	default:
		return "gagal mengklaim tiket"
	}
}

func openErrorMessage(err error) string {
	msg := err.Error()
	switch {
	case strings.Contains(msg, "P0024"):
		return "Tidak ada Tiket Hadiah. Selesaikan tugas hari ini untuk mendapatkannya."
	case strings.Contains(msg, "P0025"):
		return "Hadiah belum tersedia, coba lagi nanti"
	default:
		return "gagal membuka hadiah"
	}
}
