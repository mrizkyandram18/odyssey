package admin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/content"
	"odyssey/pkg/game/audit"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/validation"
	"odyssey/pkg/observability"
	"odyssey/pkg/shared"
)

// ContentService defines the operations the admin handler needs from the
// content caching layer.
type ContentService interface {
	Reload(ctx context.Context) error
	ReloadResource(ctx context.Context, resourceType string) error
	Status(ctx context.Context) (map[string]any, error)
	CacheStats() content.CacheStats
	CacheHitRatio() float64
	CacheGeneration() int64
	Invalidate(key string)
	ListRealms(ctx context.Context) ([]gamecontent.RealmDefinition, error)
	ListChapters(ctx context.Context) ([]gamecontent.ChapterDefinition, error)
	ListChaptersByRealm(ctx context.Context, realm string) ([]gamecontent.ChapterDefinition, error)
	ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error)
	ListQuestsByRealm(ctx context.Context, realm string) ([]gamecontent.QuestDefinition, error)
	ListPrompts(ctx context.Context) ([]gamecontent.CreativePromptDefinition, error)
	ListPromptsByRealm(ctx context.Context, realm string) ([]gamecontent.CreativePromptDefinition, error)
	ListAchievements(ctx context.Context) ([]gamecontent.AchievementDefinition, error)
	ListSeasons(ctx context.Context) ([]gamecontent.SeasonDefinition, error)
	ListLore(ctx context.Context) ([]gamecontent.LoreDefinition, error)
	ListLoreByRealm(ctx context.Context, realm string) ([]gamecontent.LoreDefinition, error)
	ListLoreByChapter(ctx context.Context, chapter string) ([]gamecontent.LoreDefinition, error)
	ListChests(ctx context.Context) ([]gamecontent.ChestDefinition, error)
	ListRelics(ctx context.Context) ([]gamecontent.RelicDefinition, error)
	GetDraft(ctx context.Context, table, slug string) (map[string]any, error)
	SaveDraft(ctx context.Context, table, slug string, patch map[string]any, updatedBy string) error
	CreateDraft(ctx context.Context, table string, data map[string]any) (map[string]any, error)
	Publish(ctx context.Context, table, slug, updatedBy string) error
	SoftDelete(ctx context.Context, table, slug string) error
	Restore(ctx context.Context, table, slug string) error
}

// AdminStore provides raw definition access for admin CRUD operations.
type AdminStore interface {
	GetByID(ctx context.Context, table string, id int64) (map[string]any, error)
	GetBySlug(ctx context.Context, table, slug string) (map[string]any, error)
	ListAll(ctx context.Context, table string, includeDeleted bool) ([]map[string]any, error)
	Create(ctx context.Context, table string, data map[string]any) (map[string]any, error)
	UpdateDraft(ctx context.Context, table, slug string, patch map[string]any, updatedBy string) error
	Publish(ctx context.Context, table, slug string, updatedBy string) error
	SoftDelete(ctx context.Context, table, slug string) error
	Restore(ctx context.Context, table, slug string) error
}

// ResourceMapping maps API resource names to DB table names.
type ResourceMapping struct {
	Table string
}

var resourceMap = map[string]ResourceMapping{
	"realms":       {gamecontent.TableRealms},
	"chapters":     {gamecontent.TableChapters},
	"quests":       {gamecontent.TableQuests},
	"prompts":      {gamecontent.TablePrompts},
	"chests":       {gamecontent.TableChests},
	"relics":       {gamecontent.TableRelics},
	"achievements": {gamecontent.TableAchievements},
	"seasons":      {gamecontent.TableSeasons},
	"lore":         {gamecontent.TableLore},
}

func getTable(resourceType string) (string, bool) {
	m, ok := resourceMap[resourceType]
	return m.Table, ok
}

// BalanceService provides runtime balance override operations.
type BalanceService interface {
	Reload(ctx context.Context) error
	Overrides() map[string]int64
	LoadedAt() time.Time
}

// AdminContext holds services wired into the admin handler.
type AdminContext struct {
	contentSvc ContentService
	adminStore AdminStore
	auditLog   *audit.Logger
	balance    BalanceService
	metrics    *observability.Metrics
}

// AdminService handles all admin CMS operations with audit logging.
type AdminService struct {
	AdminContext
}

func NewAdminService(contentSvc ContentService, adminStore AdminStore, auditLog *audit.Logger) *AdminService {
	return &AdminService{
		AdminContext: AdminContext{
			contentSvc: contentSvc,
			adminStore: adminStore,
			auditLog:   auditLog,
		},
	}
}

// SetBalance attaches an optional balance service for runtime override reload.
func (s *AdminService) SetBalance(b BalanceService) {
	s.balance = b
}

// SetMetrics attaches an optional metrics sink for validation failures.
func (s *AdminService) SetMetrics(m *observability.Metrics) {
	s.metrics = m
}

func adminUIDFromContext(ctx context.Context) string {
	uid, _ := auth.AdminUIDFromContext(ctx)
	return uid
}

func (s *AdminService) Reload(ctx context.Context) error {
	return s.contentSvc.Reload(ctx)
}

func (s *AdminService) Status(ctx context.Context) (map[string]any, error) {
	status, err := s.contentSvc.Status(ctx)
	if err != nil {
		return nil, err
	}
	status["cache_generation"] = s.contentSvc.CacheGeneration()
	status["cache_stats"] = s.contentSvc.CacheStats()
	status["cache_hit_ratio"] = s.contentSvc.CacheHitRatio()
	return status, nil
}

// ReloadBalance re-reads balance overrides from the database.
func (s *AdminService) ReloadBalance(ctx context.Context) error {
	if s.balance == nil {
		return ErrBalanceNotConfigured
	}
	return s.balance.Reload(ctx)
}

// BalanceOverrides returns the currently loaded balance overrides.
func (s *AdminService) BalanceOverrides(ctx context.Context) (map[string]any, error) {
	if s.balance == nil {
		return nil, ErrBalanceNotConfigured
	}
	overrides := s.balance.Overrides()
	out := make(map[string]any, len(overrides)+2)
	for k, v := range overrides {
		out[k] = v
	}
	out["count"] = len(overrides)
	out["loaded_at"] = s.balance.LoadedAt().UTC().Format(time.RFC3339)
	return out, nil
}

// Validate runs content validation against the live content set and records
// the result. Returns the validation result and whether the content is valid.
func (s *AdminService) Validate(ctx context.Context) (*validation.ValidationResult, error) {
	if s.contentSvc == nil {
		return nil, ErrContentNotConfigured
	}
	cs, err := buildContentSet(ctx, s.contentSvc)
	if err != nil {
		return nil, fmt.Errorf("load content: %w", err)
	}
	result := validation.NewValidator().Validate(cs)
	if !result.Valid && s.metrics != nil {
		s.metrics.RecordValidationFailure()
	}
	if s.auditLog != nil {
		_ = s.auditLog.Log(ctx, "content", "", audit.OpValidate, adminUIDFromContext(ctx), result.Valid, result)
	}
	return result, nil
}

var (
	ErrBalanceNotConfigured = errors.New("balance not configured")
	ErrContentNotConfigured = errors.New("content service not configured")
)

func (s *AdminService) List(ctx context.Context, resourceType string) ([]map[string]any, error) {
	table, ok := getTable(resourceType)
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return s.adminStore.ListAll(ctx, table, false)
}

func (s *AdminService) ListDeleted(ctx context.Context, resourceType string) ([]map[string]any, error) {
	table, ok := getTable(resourceType)
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return s.adminStore.ListAll(ctx, table, true)
}

func (s *AdminService) Get(ctx context.Context, resourceType, slug string) (map[string]any, error) {
	table, ok := getTable(resourceType)
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return s.adminStore.GetBySlug(ctx, table, slug)
}

func (s *AdminService) GetDraft(ctx context.Context, resourceType, slug string) (map[string]any, error) {
	table, ok := getTable(resourceType)
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	return s.contentSvc.GetDraft(ctx, table, slug)
}

func (s *AdminService) Create(ctx context.Context, resourceType string, data map[string]any) (map[string]any, error) {
	table, ok := getTable(resourceType)
	if !ok {
		return nil, fmt.Errorf("unknown resource type: %s", resourceType)
	}
	created, err := s.adminStore.Create(ctx, table, data)
	if err != nil {
		return nil, err
	}
	if s.auditLog != nil {
		uid := adminUIDFromContext(ctx)
		slug := ""
		if sl, ok := created["slug"].(string); ok {
			slug = sl
		}
		_ = s.auditLog.Log(ctx, resourceType, slug, audit.OpCreate, uid, nil, created)
	}
	_ = s.contentSvc.ReloadResource(ctx, resourceType)
	return created, nil
}

func (s *AdminService) SaveDraft(ctx context.Context, resourceType, slug string, patch map[string]any) error {
	table, ok := getTable(resourceType)
	if !ok {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	uid := adminUIDFromContext(ctx)
	if s.auditLog != nil {
		oldVal, _ := s.adminStore.GetBySlug(ctx, table, slug)
		_ = s.auditLog.Log(ctx, resourceType, slug, audit.OpSaveDraft, uid, oldVal, patch)
	}
	if err := s.adminStore.UpdateDraft(ctx, table, slug, patch, uid); err != nil {
		return err
	}
	_ = s.contentSvc.ReloadResource(ctx, resourceType)
	return nil
}

func (s *AdminService) Publish(ctx context.Context, resourceType, slug string) error {
	table, ok := getTable(resourceType)
	if !ok {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	uid := adminUIDFromContext(ctx)
	if s.auditLog != nil {
		oldVal, _ := s.adminStore.GetBySlug(ctx, table, slug)
		_ = s.auditLog.Log(ctx, resourceType, slug, audit.OpPublish, uid, oldVal, nil)
	}
	if err := s.adminStore.Publish(ctx, table, slug, uid); err != nil {
		return err
	}
	_ = s.contentSvc.ReloadResource(ctx, resourceType)
	return nil
}

func (s *AdminService) Delete(ctx context.Context, resourceType, slug string) error {
	table, ok := getTable(resourceType)
	if !ok {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	uid := adminUIDFromContext(ctx)
	if s.auditLog != nil {
		oldVal, _ := s.adminStore.GetBySlug(ctx, table, slug)
		_ = s.auditLog.Log(ctx, resourceType, slug, audit.OpDelete, uid, oldVal, nil)
	}
	if err := s.adminStore.SoftDelete(ctx, table, slug); err != nil {
		return err
	}
	_ = s.contentSvc.ReloadResource(ctx, resourceType)
	return nil
}

func (s *AdminService) Restore(ctx context.Context, resourceType, slug string) error {
	table, ok := getTable(resourceType)
	if !ok {
		return fmt.Errorf("unknown resource type: %s", resourceType)
	}
	uid := adminUIDFromContext(ctx)
	if s.auditLog != nil {
		_ = s.auditLog.Log(ctx, resourceType, slug, audit.OpRestore, uid, nil, nil)
	}
	if err := s.adminStore.Restore(ctx, table, slug); err != nil {
		return err
	}
	_ = s.contentSvc.ReloadResource(ctx, resourceType)
	return nil
}

var svc *AdminService

func Setup(s *AdminService) {
	svc = s
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if !auth.IsAdmin(r) {
		shared.WriteUnauthorized(w)
		return
	}
	if svc == nil {
		shared.WriteJSONError(w, "server not configured", http.StatusServiceUnavailable)
		return
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	path := trimAdminPath(r.URL.Path)
	parts := strings.Split(path, "/")
	parts = filterEmpty(parts)

	if len(parts) == 0 {
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
		return
	}

	first := parts[0]
	if first == "reload" && r.Method == http.MethodPost {
		handleReload(w, r)
		return
	}
	if first == "status" && r.Method == http.MethodGet {
		handleStatus(w, r)
		return
	}
	if first == "metrics" && r.Method == http.MethodGet {
		handleMetrics(w, r)
		return
	}
	if first == "validate" && r.Method == http.MethodGet {
		handleValidate(w, r)
		return
	}
	if first == "audit" && r.Method == http.MethodGet {
		handleAuditList(w, r)
		return
	}
	if first == "balance" && r.Method == http.MethodGet {
		handleBalanceList(w, r)
		return
	}
	if first == "balance" && r.Method == http.MethodPost {
		handleBalanceSet(w, r)
		return
	}

	if len(parts) < 2 {
		shared.WriteJSONError(w, "not found", http.StatusNotFound)
		return
	}

	resourceType := parts[0]
	if _, ok := getTable(resourceType); !ok {
		shared.WriteJSONError(w, "unknown resource type: "+resourceType, http.StatusBadRequest)
		return
	}

	if len(parts) == 2 {
		switch r.Method {
		case http.MethodGet:
			if parts[1] == "deleted" {
				handleListDeleted(w, r, resourceType)
			} else {
				handleList(w, r, resourceType)
			}
		case http.MethodPost:
			if parts[1] == "" {
				handleCreate(w, r, resourceType)
			} else {
				shared.WriteJSONError(w, "not found", http.StatusNotFound)
			}
		default:
			shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	if len(parts) >= 3 {
		slug := parts[1]
		action := parts[2]

		switch {
		case r.Method == http.MethodGet && slug == "draft" && action != "":
			shared.WriteJSONError(w, "invalid path", http.StatusBadRequest)
		case r.Method == http.MethodGet && parts[1] != "" && len(parts) == 2:
			handleGet(w, r, resourceType, parts[1])
		case r.Method == http.MethodGet && action == "draft":
			if len(parts) >= 4 {
				handleGetDraft(w, r, resourceType, parts[3])
			} else {
				shared.WriteJSONError(w, "slug required", http.StatusBadRequest)
			}
		case r.Method == http.MethodPost && action == "draft":
			if len(parts) >= 4 {
				handleSaveDraft(w, r, resourceType, parts[3])
			} else {
				shared.WriteJSONError(w, "slug required", http.StatusBadRequest)
			}
		case r.Method == http.MethodPatch && action == "draft":
			if len(parts) >= 4 {
				handleSaveDraft(w, r, resourceType, parts[3])
			} else {
				shared.WriteJSONError(w, "slug required", http.StatusBadRequest)
			}
		case r.Method == http.MethodPatch && action == "publish":
			if len(parts) >= 4 {
				handlePublish(w, r, resourceType, parts[3])
			} else {
				shared.WriteJSONError(w, "slug required", http.StatusBadRequest)
			}
		case r.Method == http.MethodPatch && action == "delete":
			if len(parts) >= 4 {
				handleDelete(w, r, resourceType, parts[3])
			} else {
				shared.WriteJSONError(w, "slug required", http.StatusBadRequest)
			}
		case r.Method == http.MethodPatch && action == "restore":
			if len(parts) >= 4 {
				handleRestore(w, r, resourceType, parts[3])
			} else {
				shared.WriteJSONError(w, "slug required", http.StatusBadRequest)
			}
		case r.Method == http.MethodGet:
			handleGet(w, r, resourceType, slug)
		default:
			shared.WriteJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}

	shared.WriteJSONError(w, "not found", http.StatusNotFound)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	status, err := svc.Status(r.Context())
	if err != nil {
		shared.WriteJSONError(w, "failed to get content status", http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, status)
}

func handleReload(w http.ResponseWriter, r *http.Request) {
	err := svc.Reload(r.Context())
	if err != nil {
		shared.WriteJSONError(w, "failed to reload content: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{
		"status":    "reloaded",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	_ = r
	stats := svc.contentSvc.CacheStats()
	metrics := map[string]any{
		"cache_hits":       stats.Hits,
		"cache_misses":     stats.Misses,
		"cache_evictions":  stats.Evictions,
		"cache_hit_ratio":  svc.contentSvc.CacheHitRatio(),
		"cache_generation": svc.contentSvc.CacheGeneration(),
		"timestamp":        time.Now().UTC().Format(time.RFC3339),
	}
	shared.WriteJSON(w, http.StatusOK, metrics)
}

func handleAuditList(w http.ResponseWriter, r *http.Request) {
	_ = r
	shared.WriteJSON(w, http.StatusNotImplemented, map[string]string{"error": "audit log not configured"})
}

func handleBalanceList(w http.ResponseWriter, r *http.Request) {
	overrides, err := svc.BalanceOverrides(r.Context())
	if err != nil {
		shared.WriteJSONError(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	shared.WriteJSON(w, http.StatusOK, overrides)
}

func handleBalanceSet(w http.ResponseWriter, r *http.Request) {
	uid := adminUIDFromContext(r.Context())
	if svc.auditLog != nil {
		_ = svc.auditLog.Log(r.Context(), "balance", "", audit.OpReload, uid, nil, nil)
	}
	if err := svc.ReloadBalance(r.Context()); err != nil {
		shared.WriteJSONError(w, "failed to reload balance: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{
		"status":    "reloaded",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func handleValidate(w http.ResponseWriter, r *http.Request) {
	result, err := svc.Validate(r.Context())
	if err != nil {
		shared.WriteJSONError(w, "failed to validate: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, result)
}

// buildContentSet assembles a validation.ContentSet from the content service.
func buildContentSet(ctx context.Context, cs ContentService) (validation.ContentSet, error) {
	out := validation.ContentSet{}
	var err error
	out.Realms, err = cs.ListRealms(ctx)
	if err != nil {
		return out, err
	}
	out.Chapters, err = cs.ListChapters(ctx)
	if err != nil {
		return out, err
	}
	out.Quests, err = cs.ListQuests(ctx)
	if err != nil {
		return out, err
	}
	out.Prompts, err = cs.ListPrompts(ctx)
	if err != nil {
		return out, err
	}
	out.Achievements, err = cs.ListAchievements(ctx)
	if err != nil {
		return out, err
	}
	out.Seasons, err = cs.ListSeasons(ctx)
	if err != nil {
		return out, err
	}
	out.Lore, err = cs.ListLore(ctx)
	if err != nil {
		return out, err
	}
	out.Chests, err = cs.ListChests(ctx)
	if err != nil {
		return out, err
	}
	out.Relics, err = cs.ListRelics(ctx)
	if err != nil {
		return out, err
	}
	return out, nil
}

func handleList(w http.ResponseWriter, r *http.Request, resourceType string) {
	defs, err := svc.List(r.Context(), resourceType)
	if err != nil {
		shared.WriteJSONError(w, "failed to list "+resourceType+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, defs)
}

func handleListDeleted(w http.ResponseWriter, r *http.Request, resourceType string) {
	defs, err := svc.ListDeleted(r.Context(), resourceType)
	if err != nil {
		shared.WriteJSONError(w, "failed to list deleted "+resourceType+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, defs)
}

func handleGet(w http.ResponseWriter, r *http.Request, resourceType, slug string) {
	if !shared.ValidateSlug(slug) {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	def, err := svc.Get(r.Context(), resourceType, slug)
	if err != nil {
		shared.WriteJSONError(w, "failed to get "+resourceType+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	if def == nil {
		shared.WriteJSONError(w, resourceType+" not found", http.StatusNotFound)
		return
	}
	shared.WriteJSON(w, http.StatusOK, def)
}

func handleGetDraft(w http.ResponseWriter, r *http.Request, resourceType, slug string) {
	if !shared.ValidateSlug(slug) {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	def, err := svc.GetDraft(r.Context(), resourceType, slug)
	if err != nil {
		shared.WriteJSONError(w, "failed to get draft: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if def == nil {
		shared.WriteJSONError(w, resourceType+" draft not found", http.StatusNotFound)
		return
	}
	shared.WriteJSON(w, http.StatusOK, def)
}

func handleCreate(w http.ResponseWriter, r *http.Request, resourceType string) {
	var data map[string]any
	if err := shared.ReadJSON(r, &data); err != nil {
		shared.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(data) > 100 {
		shared.WriteJSONError(w, "payload too large", http.StatusBadRequest)
		return
	}
	created, err := svc.Create(r.Context(), resourceType, data)
	if err != nil {
		shared.WriteJSONError(w, "failed to create "+resourceType+": "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusCreated, created)
}

func handleSaveDraft(w http.ResponseWriter, r *http.Request, resourceType, slug string) {
	if !shared.ValidateSlug(slug) {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	var patch map[string]any
	if err := shared.ReadJSON(r, &patch); err != nil {
		shared.WriteJSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	if len(patch) > 100 {
		shared.WriteJSONError(w, "patch too large", http.StatusBadRequest)
		return
	}
	if err := svc.SaveDraft(r.Context(), resourceType, slug, patch); err != nil {
		shared.WriteJSONError(w, "failed to save draft: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "draft_saved"})
}

func handlePublish(w http.ResponseWriter, r *http.Request, resourceType, slug string) {
	if !shared.ValidateSlug(slug) {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	_ = r
	if err := svc.Publish(r.Context(), resourceType, slug); err != nil {
		shared.WriteJSONError(w, "failed to publish: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "published"})
}

func handleDelete(w http.ResponseWriter, r *http.Request, resourceType, slug string) {
	if !shared.ValidateSlug(slug) {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	_ = r
	if err := svc.Delete(r.Context(), resourceType, slug); err != nil {
		shared.WriteJSONError(w, "failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func handleRestore(w http.ResponseWriter, r *http.Request, resourceType, slug string) {
	if !shared.ValidateSlug(slug) {
		shared.WriteJSONError(w, "invalid slug", http.StatusBadRequest)
		return
	}
	_ = r
	if err := svc.Restore(r.Context(), resourceType, slug); err != nil {
		shared.WriteJSONError(w, "failed to restore: "+err.Error(), http.StatusInternalServerError)
		return
	}
	shared.WriteJSON(w, http.StatusOK, map[string]string{"status": "restored"})
}

func trimAdminPath(path string) string {
	p := path
	for len(p) > 0 && p[0] == '/' {
		p = p[1:]
	}
	if len(p) >= 6 && p[:6] == "admin/" {
		p = p[6:]
	}
	return p
}

func filterEmpty(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func init() {
	_ = strconv.Itoa(0)
	_ = time.Now
}
