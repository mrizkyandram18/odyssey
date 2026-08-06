package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	apiAchievements "odyssey/api/achievements"
	"odyssey/api/admin"
	apiChapters "odyssey/api/chapters"
	"odyssey/api/chests"
	"odyssey/api/creative"
	"odyssey/api/crews"
	"odyssey/api/daily_turns"
	apiHome "odyssey/api/home"
	"odyssey/api/login"
	apiLore "odyssey/api/lore"
	"odyssey/api/me"
	"odyssey/api/quests"
	apiReactions "odyssey/api/reactions"
	"odyssey/api/realm_progress"
	"odyssey/api/relics"
	"odyssey/api/status"
	"odyssey/pkg/auth"
	"odyssey/pkg/content"
	"odyssey/pkg/db"
	"odyssey/pkg/game/achievement"
	"odyssey/pkg/game/audit"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/chapter"
	"odyssey/pkg/game/chest"
	gamecreative "odyssey/pkg/game/creative"
	"odyssey/pkg/game/dailyturn"
	"odyssey/pkg/game/events"
	gamehome "odyssey/pkg/game/home"
	"odyssey/pkg/game/lore"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/quest"
	"odyssey/pkg/game/relic"
	"odyssey/pkg/game/season"
	"odyssey/pkg/game/social"
	"odyssey/pkg/game/world"
	"odyssey/pkg/observability"
	"odyssey/pkg/shared"
)

func main() {
	_ = godotenv.Load("../../.env")

	config := shared.LoadConfig()
	if err := config.Validate(); err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	ctx := context.Background()

	store, err := auth.NewFirestoreClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	firestoreReader := auth.NewFirestoreStore(store)

	supabaseClient := db.NewClient(config.SupabaseURL, config.SupabaseServiceKey)
	profileStore := db.NewProfileStore(supabaseClient)
	repo, err := db.BuildRepository(supabaseClient)
	if err != nil {
		log.Fatalf("Failed to build repository: %v", err)
	}

	realmCfg := world.DefaultRealmCatalog
	if err := realmCfg.ApplyOverrides(ctx, repo.Config); err != nil {
		log.Printf("Warning: realm config overrides failed: %v", err)
	}
	progCfg := progression.DefaultProgressionConfig()
	realmStore := db.NewRealmProgressStore(supabaseClient)
	chapterStore := db.NewChapterProgressStore(supabaseClient)

	contentRepo, err := db.BuildContentRepository(supabaseClient)
	if err != nil {
		log.Fatalf("Failed to build content repository: %v", err)
	}

	contentSvc := content.NewContentService(
		contentRepo.Realms,
		contentRepo.Chapters,
		contentRepo.Quests,
		contentRepo.Prompts,
		contentRepo.Achievements,
		contentRepo.Seasons,
		contentRepo.Lore,
		repo.ChestDefinitions,
		repo.RelicDefinitions,
		content.ContentServiceConfig{CacheTTL: 5 * time.Minute},
	)
	contentSvc.SetAdminStore(db.NewDefinitionStore(supabaseClient))

	if err := contentSvc.Reload(ctx); err != nil {
		log.Printf("Warning: initial content reload failed: %v", err)
	}

	dispatcher := events.NewDispatcher()

	metrics := observability.NewMetrics()
	profiler := observability.NewProfiler()
	logger := observability.DefaultLogger()
	dispatcher.SetRecorder(metrics)

	schemaVer, readErr := observability.ReadSchemaVersion(ctx, supabaseClient)
	if readErr != nil {
		logger.Warn("schema_version_read_failed", map[string]any{
			"error": readErr.Error(),
		})
	} else {
		observability.SetRuntimeSchemaVersion(schemaVer)
		if schemaVer != observability.SchemaVersion {
			logger.Warn("schema_version_mismatch", map[string]any{
				"build_version": observability.SchemaVersion,
				"db_version":    schemaVer,
			})
		} else {
			logger.Info("schema_version_loaded", map[string]any{
				"schema_version": schemaVer,
			})
		}
	}

	go syncCacheStats(ctx, contentSvc, metrics)

	seasonSvc := season.NewSeasonService(contentRepo.Seasons, nil)

	questSvc := quest.NewQuestServiceWithGate(
		repo.Quests,
		quest.NewQuestGate(chapterStore, realmStore, repo.Users, repo.Quests, seasonSvc.IsActive),
		contentSvc,
	)
	questSvc.SetUserStore(repo.Users)

	chapterSvc := chapter.NewChapterService(chapterStore, repo.Quests, contentSvc, dispatcher)
	chapterSvc.SetMetrics(metrics)
	loreSvc := lore.NewLoreServiceWithPublisher(repo.LoreUnlocks, contentSvc, dispatcher)
	achieveRdr := achievement.NewProgressReader(
		repo.Quests, realmStore, repo.Users,
		repo.PlayerRelics, repo.DailyTurns, repo.CreativeSubmissions, chapterStore,
	)
	achieveSvc := achievement.NewAchievementService(contentRepo.Achievements, repo.Achievements, achieveRdr, dispatcher)

	dispatcher.Subscribe(events.EventTypeQuestCompleted, chapter.NewQuestCompletedHandler(chapterSvc))
	dispatcher.Subscribe(events.EventTypeQuestCompleted, achievement.NewQuestCompletedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeChapterCompleted, lore.NewChapterCompletedHandler(loreSvc))
	dispatcher.Subscribe(events.EventTypeChapterCompleted, achievement.NewChapterCompletedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeRelicCollected, achievement.NewRelicCollectedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeDailyTurnCompleted, achievement.NewDailyTurnCompletedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeLevelReached, achievement.NewLevelReachedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeCreativeSubmission, achievement.NewCreativeSubmissionHandler(achieveSvc))

	creativeSvc := gamecreative.NewCreativeServiceWithPublisher(repo.CreativeSubmissions, repo.Quests, dispatcher)
	creativeHandler := gamecreative.NewCreativeAPIHandler(creativeSvc)

	dailyTurnCfg := dailyturn.DailyTurnConfig{
		XP:             config.DailyTurnXP,
		MaxTurnsPerDay: config.MaxDailyTurnsPerDay,
		Timezone:       config.Timezone,
	}
	dailyTurnSvc := dailyturn.NewDailyTurnService(repo.DailyTurns, &dailyTurnCfg)
	dailyTurnSvc.SetMetrics(metrics)

	balanceSvc := balance.NewService(db.NewBalanceStore(supabaseClient))
	if err := balanceSvc.Load(ctx); err != nil {
		log.Printf("Warning: balance load failed: %v", err)
	}

	progSvc := progression.NewProgressionServiceWithPublisher(repo.Users, &progCfg, dispatcher, balanceSvc)
	progSvc.SetMetrics(metrics)

	questHandler := quest.NewQuestAPIHandler(questSvc, progSvc, realmStore, realmCfg, &progCfg)
	questHandler.SetPublisher(dispatcher)
	questHandler.SetContentGateway(contentSvc)
	questHandler.SetBalance(balanceSvc)
	questHandler.SetMetrics(metrics)
	questHandler.SetLogger(logger)
	dailyTurnHandler := dailyturn.NewDailyTurnAPIHandlerWithPublisher(dailyTurnSvc, progSvc, dispatcher, balanceSvc)
	dailyTurnHandler.SetMetrics(metrics)
	dailyTurnHandler.SetLogger(logger)
	dailyTurnHandler.SetActivityStore(repo.Activity)

	chestEngine := chest.NewRewardEngine(repo.ChestDefinitions, repo.RelicDefinitions)
	chestSvc := chest.NewChestServiceWithPublisher(repo.Chests, repo.PlayerRelics, repo.Relics, repo.Users, chestEngine, dispatcher)
	chestSvc.SetContentGateway(contentSvc)
	chestSvc.SetBalance(balanceSvc)
	chestSvc.SetMetrics(metrics)
	chests.Setup(chestSvc)
	dispatcher.Subscribe(events.EventTypeQuestCompleted, chest.NewQuestCompletedHandler(chestSvc, contentSvc))

	relicSvc := relic.NewRelicService(repo.Relics, repo.PlayerRelics, repo.RelicDefinitions)
	relics.Setup(relicSvc)

	homeSvc := gamehome.NewHomeService(questSvc, dailyTurnSvc, repo.Progression, realmStore, repo.Users, repo.CreativeSubmissions, repo.Chests, relicSvc)
	homeSvc.SetChapterService(chapterSvc)
	homeSvc.SetLoreService(loreSvc)
	homeSvc.SetAchievementService(achieveSvc)

	adminStore := db.NewDefinitionStore(supabaseClient)
	auditStore := db.NewAuditStore(supabaseClient)
	auditLogger := audit.NewLogger(auditStore)
	adminSvc := admin.NewAdminService(contentSvc, adminStore, auditLogger)
	adminSvc.SetBalance(balanceSvc)
	adminSvc.SetMetrics(metrics)
	admin.Setup(adminSvc)

	authenticator := auth.NewFirestoreAuthenticator(
		config.ParentID,
		config.MinBuildNumber,
		auth.NewBcryptHasher(),
		firestoreReader,
		profileStore,
	)

	sessionSecret := config.SessionSecret
	if sessionSecret == "" {
		sessionSecret = config.AdminSecret
	}
	issuer := auth.NewSessionIssuer(sessionSecret)
	mw := auth.NewMiddleware(issuer)

	login.Setup(authenticator, issuer, profileStore)
	me.Setup(profileStore)
	quests.Setup(questHandler)
	daily_turns.Setup(dailyTurnHandler)
	creative.Setup(creativeHandler)
	apiHome.Setup(homeSvc)

	apiChapters.Setup(chapterSvc)
	apiLore.Setup(loreSvc)
	apiAchievements.Setup(achieveSvc)

	reactionSvc := social.NewReactionService(repo.Reactions)
	apiReactions.Setup(reactionSvc)

	crews.Setup(repo.Crews)
	realm_progress.Setup(repo.RealmProgress)

	secCfg := shared.DefaultSecurityConfig()
	secCfg.AllowedOrigins = config.AllowedOrigins
	secCfg.MaxBodyBytes = config.MaxBodyBytes
	secCfg.RateLimitWindow = time.Duration(config.RateLimitWindowSec) * time.Second
	secCfg.RateLimitMaxHits = config.RateLimitMaxHits
	secCfg.LoginRateLimitMax = config.LoginRateLimitMax
	secCfg.AdminRateLimitMax = config.AdminRateLimitMax

	userLimiter := shared.NewRateLimiter(secCfg.RateLimitWindow, secCfg.RateLimitMaxHits)
	loginLimiter := shared.NewRateLimiter(secCfg.RateLimitWindow, secCfg.LoginRateLimitMax)
	adminLimiter := shared.NewRateLimiter(secCfg.RateLimitWindow, secCfg.AdminRateLimitMax)

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			userLimiter.Cleanup()
			loginLimiter.Cleanup()
			adminLimiter.Cleanup()
		}
	}()

	hc := observability.NewHealthChecker(
		observability.HealthCheck{Name: "configuration", Fn: func(ctx context.Context) error {
			return config.Validate()
		}},
		observability.HealthCheck{Name: "database", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_system_config")},
		observability.HealthCheck{Name: "cache", Fn: func(ctx context.Context) error {
			if contentSvc == nil {
				return errors.New("content service not initialized")
			}
			return nil
		}},
		observability.HealthCheck{Name: "content_generation", Fn: func(ctx context.Context) error {
			if contentSvc == nil {
				return errors.New("content service unavailable")
			}
			_, err := contentSvc.Status(ctx)
			return err
		}},
		observability.HealthCheck{Name: "audit_store", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_audit_logs")},
		observability.HealthCheck{Name: "admin_store", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_realm_definitions")},
	)

	status.Setup(status.FuncStatusProvider(func(ctx context.Context) map[string]any {
		info := map[string]any{
			"app":            "odyssey",
			"version":        observability.Version,
			"schema_version": observability.SchemaVersionString(),
			"uptime_seconds": int64(time.Since(metrics.BootTime()).Seconds()),
			"boot_time":      metrics.BootTime().UTC().Format(time.RFC3339),
		}
		if contentSvc != nil {
			if st, err := contentSvc.Status(ctx); err == nil {
				info["content_status"] = st
				info["content_generation"] = contentSvc.CacheGeneration()
				info["cache_hit_ratio"] = contentSvc.CacheHitRatio()
			}
		}
		return info
	}))

	mux := http.NewServeMux()

	secure := func(next http.HandlerFunc) http.HandlerFunc {
		return shared.SecurityHeadersMiddleware(secCfg, shared.CORSHeaderMiddleware(secCfg, shared.RequestLimitMiddleware(secCfg, next)))
	}

	csrf := func(next http.HandlerFunc) http.HandlerFunc {
		return shared.CSRFMiddleware(secCfg, next)
	}

	rateLimit := func(limiter *shared.RateLimiter, next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := r.RemoteAddr
			if !limiter.Allow(key) {
				shared.WriteJSONError(w, "rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		}
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "../../web/dist/index.html")
	})

	mux.HandleFunc("/api/login", secure(rateLimit(loginLimiter, login.Handler)))
	mux.HandleFunc("/api/login/", secure(rateLimit(loginLimiter, login.Handler)))
	mux.HandleFunc("/api/csrf", secure(func(w http.ResponseWriter, r *http.Request) {
		token := shared.GenerateCSRFToken()
		http.SetCookie(w, &http.Cookie{
			Name:     "odyssey_csrf",
			Value:    token,
			Path:     "/",
			MaxAge:   secCfg.CSRFMaxAge,
			HttpOnly: false,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		})
		shared.WriteJSON(w, http.StatusOK, map[string]string{"csrf_token": token})
	}))
	mux.HandleFunc("/api/me", secure(mw.RequireAuth(me.Handler)))
	mux.HandleFunc("/api/me/", secure(mw.RequireAuth(me.Handler)))
	mux.HandleFunc("/api/status", secure(status.Handler))
	mux.HandleFunc("/api/status/", secure(status.Handler))
	mux.HandleFunc("/api/quests", secure(mw.RequireAuth(quests.Handler)))
	mux.HandleFunc("/api/quests/", secure(mw.RequireAuth(quests.Handler)))
	mux.HandleFunc("/api/crews", secure(mw.RequireAuth(crews.Handler)))
	mux.HandleFunc("/api/crews/", secure(mw.RequireAuth(crews.Handler)))
	mux.HandleFunc("/api/realm_progress", secure(mw.RequireAuth(realm_progress.Handler)))
	mux.HandleFunc("/api/realm_progress/", secure(mw.RequireAuth(realm_progress.Handler)))
	mux.HandleFunc("/api/daily_turns", secure(mw.RequireAuth(daily_turns.Handler)))
	mux.HandleFunc("/api/daily_turns/", secure(mw.RequireAuth(daily_turns.Handler)))
	mux.HandleFunc("/api/creative", secure(mw.RequireAuth(csrf(creative.Handler))))
	mux.HandleFunc("/api/creative/", secure(mw.RequireAuth(csrf(creative.Handler))))
	mux.HandleFunc("/api/home", secure(mw.RequireAuth(apiHome.Handler)))
	mux.HandleFunc("/api/home/", secure(mw.RequireAuth(apiHome.Handler)))
	mux.HandleFunc("/api/chests", secure(mw.RequireAuth(chests.Handler)))
	mux.HandleFunc("/api/chests/", secure(mw.RequireAuth(chests.Handler)))
	mux.HandleFunc("/api/relics", secure(mw.RequireAuth(relics.Handler)))
	mux.HandleFunc("/api/relics/", secure(mw.RequireAuth(relics.Handler)))
	mux.HandleFunc("/api/chapters", secure(mw.RequireAuth(apiChapters.Handler)))
	mux.HandleFunc("/api/chapters/", secure(mw.RequireAuth(apiChapters.Handler)))
	mux.HandleFunc("/api/lore", secure(mw.RequireAuth(apiLore.Handler)))
	mux.HandleFunc("/api/lore/", secure(mw.RequireAuth(apiLore.Handler)))
	mux.HandleFunc("/api/achievements", secure(mw.RequireAuth(apiAchievements.Handler)))
	mux.HandleFunc("/api/achievements/", secure(mw.RequireAuth(apiAchievements.Handler)))
	mux.HandleFunc("/api/reactions", secure(mw.RequireAuth(apiReactions.Handler)))
	mux.HandleFunc("/api/reactions/", secure(mw.RequireAuth(apiReactions.Handler)))
	mux.HandleFunc("/api/admin", secure(rateLimit(adminLimiter, mw.RequireRole(auth.RoleAdmin)(admin.Handler))))
	mux.HandleFunc("/api/admin/", secure(rateLimit(adminLimiter, mw.RequireRole(auth.RoleAdmin)(admin.Handler))))

	mux.HandleFunc("/metrics", secure(observability.InternalTokenMiddleware(config.InternalMetricsToken, observability.MetricsHandler(metrics))))
	mux.HandleFunc("/version", secure(observability.VersionHandler(&contentSvcGenProvider{svc: contentSvc})))
	mux.HandleFunc("/health", secure(observability.HealthHandler(hc)))
	mux.HandleFunc("/ready", secure(observability.ReadyHandler(hc)))
	mux.HandleFunc("/live", secure(observability.LiveHandler(hc)))
	mux.HandleFunc("/debug/profile", secure(observability.InternalTokenMiddleware(config.InternalMetricsToken, observability.ProfileHandler(profiler))))
	mux.HandleFunc("/debug/profile/recommendations", secure(observability.InternalTokenMiddleware(config.InternalMetricsToken, observability.RecommendationsHandler(profiler))))

	obs := &observability.Observability{
		Logger:   logger,
		Metrics:  metrics,
		Profiler: profiler,
		Health:   hc,
	}

	handler := obs.Wrap(mux)
	srv := &http.Server{
		Addr:              ":" + getPort(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	port := getPort()
	log.Printf("Odyssey server starting on :%s", port)
	log.Printf("Timezone: %s | DailyTurnXP: %d | MaxDailyTurns: %d | RealmCatalog: %d realms",
		config.Timezone, config.DailyTurnXP, config.MaxDailyTurnsPerDay, len(realmCfg.Order()))
	log.Printf("Routes: /api/login /api/me /api/status /api/quests /api/crews /api/realm_progress /api/daily_turns /api/creative /api/home /api/chests /api/relics /api/chapters /api/lore /api/achievements /api/admin /metrics /version /health /ready /live /debug/profile")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sm := observability.NewShutdownManager(srv, 30*time.Second)
	sm.AddHook("content_service_stop", func(ctx context.Context) error {
		contentSvc.Stop()
		return nil
	})
	sm.AddHook("dispatcher_close", func(ctx context.Context) error {
		dispatcher.Close()
		return nil
	})
	sm.AddHook("rate_limiter_cleanup", func(ctx context.Context) error {
		userLimiter.Cleanup()
		loginLimiter.Cleanup()
		adminLimiter.Cleanup()
		return nil
	})
	sm.AddHook("logger_close", func(ctx context.Context) error {
		logger.Close()
		return nil
	})
	sm.AddHook("cache_stats_stop", func(ctx context.Context) error {
		// The syncCacheStats goroutine will stop when the main context is cancelled
		return nil
	})

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()

	logger.Info("server_started", map[string]any{
		"port":    port,
		"version": observability.Version,
	})

	<-sigCh
	logger.Info("shutdown_signal_received", nil)
	if err := sm.Shutdown(ctx); err != nil {
		log.Fatalf("Graceful shutdown error: %v", err)
	}
	logger.Info("server_stopped", nil)
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return port
}

func syncCacheStats(ctx context.Context, svc *content.ContentService, metrics *observability.Metrics) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			stats := svc.CacheStats()
			metrics.MergeCacheStats(stats.Hits, stats.Misses, 0)
		}
	}
}

type contentSvcGenProvider struct {
	svc *content.ContentService
}

func (p *contentSvcGenProvider) GetContentGeneration() int64 {
	if p == nil || p.svc == nil {
		return 0
	}
	return p.svc.CacheGeneration()
}
