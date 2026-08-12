package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	apiAchievements "odyssey/internal/api/achievements"
	"odyssey/internal/api/admin"
	apiBoard "odyssey/internal/api/board"
	apiChapters "odyssey/internal/api/courses"
	"odyssey/internal/api/gifts"
	apiCosmetics "odyssey/internal/api/cosmetics"
	"odyssey/internal/api/creative"
	"odyssey/internal/api/families"
	"odyssey/internal/api/daily_activities"
	"odyssey/internal/api/daily_missions"
	apiHome "odyssey/internal/api/home"
	"odyssey/internal/api/login"
	apiLore "odyssey/internal/api/concepts"
	"odyssey/internal/api/me"
	apiPush "odyssey/internal/api/push"
	apiQuests "odyssey/internal/api/missions"
	apiReactions "odyssey/internal/api/reactions"
	"odyssey/internal/api/journey_progress"
	"odyssey/internal/api/collections"
	"odyssey/internal/api/rewards"
	apiSeasons "odyssey/internal/api/seasons"
	"odyssey/internal/api/status"
	apiStoryFragments "odyssey/internal/api/story_fragments"
	"odyssey/pkg/auth"
	"odyssey/pkg/content"
	"odyssey/pkg/db"
	gameadmin "odyssey/pkg/game/admin"
	"odyssey/pkg/game/dailyactivity"
	"odyssey/pkg/game/achievement"
	"odyssey/pkg/game/audit"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/board"
	"odyssey/pkg/game/course"
	"odyssey/pkg/game/gift"
	"odyssey/pkg/game/cosmetic"
	gamecreative "odyssey/pkg/game/creative"
	"odyssey/pkg/game/familystreak"
	"odyssey/pkg/game/dailymission"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/fragment"
	gamehome "odyssey/pkg/game/home"
	"odyssey/pkg/game/concepts"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/mission"
	"odyssey/pkg/game/collection"
	"odyssey/pkg/game/reward"
	"odyssey/pkg/game/rewardsignal"
	"odyssey/pkg/game/season"
	"odyssey/pkg/game/social"
	"odyssey/pkg/game/world"
	"odyssey/pkg/observability"
	"odyssey/pkg/push"
	"odyssey/pkg/shared"
)

type Server struct {
	Handler http.Handler
	Cleanup func(context.Context) error
}

func BuildHandler() (*Server, error) {
	config := shared.LoadConfig()
	if err := config.Validate(); err != nil {
		return nil, err
	}

	ctx := context.Background()

	supabaseClient := db.NewClient(config.SupabaseURL, config.SupabaseServiceKey)
	profileStore := db.NewProfileStore(supabaseClient)
	repo, err := db.BuildRepository(supabaseClient)
	if err != nil {
		return nil, errors.New("Failed to build repository: " + err.Error())
	}

	realmCfg := world.DefaultRealmCatalog
	if err := realmCfg.ApplyOverrides(ctx, repo.Config); err != nil {
		log.Printf("Warning: journey config overrides failed: %v", err)
	}
	progCfg := progression.DefaultProgressionConfig()
	realmStore := db.NewRealmProgressStore(supabaseClient)
	chapterStore := db.NewChapterProgressStore(supabaseClient)

	contentRepo, err := db.BuildContentRepository(supabaseClient)
	if err != nil {
		return nil, errors.New("Failed to build content repository: " + err.Error())
	}

	contentSvc := content.NewContentService(
		contentRepo.Journeys,
		contentRepo.Courses,
		contentRepo.Missions,
		contentRepo.Prompts,
		contentRepo.Achievements,
		contentRepo.Seasons,
		contentRepo.Concept,
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

	// We pass a new context for background tasks so they can be cancelled on shutdown
	bgCtx, cancelBg := context.WithCancel(context.Background())
	go syncCacheStats(bgCtx, contentSvc, metrics)

	seasonSvc := season.NewSeasonService(contentRepo.Seasons, nil)

	questSvc := quest.NewQuestServiceWithGate(
		repo.Missions,
		quest.NewQuestGate(chapterStore, realmStore, repo.Users, repo.Missions, seasonSvc.IsActive),
		contentSvc,
	)
	questSvc.SetUserStore(repo.Users)

	chapterSvc := course.NewChapterService(chapterStore, repo.Missions, contentSvc, dispatcher)
	chapterSvc.SetMetrics(metrics)
	loreSvc := concept.NewLoreServiceWithPublisher(repo.LoreUnlocks, contentSvc, dispatcher)
	achieveRdr := achievement.NewProgressReader(
		repo.Missions, realmStore, repo.Users,
		repo.PlayerRelics, repo.DailyTurns, repo.CreativeSubmissions, chapterStore,
	)
	achieveSvc := achievement.NewAchievementService(contentRepo.Achievements, repo.Achievements, achieveRdr, dispatcher, repo.CosmeticUnlocks)
	rewardSignalService := rewardsignal.NewService(repo)

	dispatcher.Subscribe(events.EventTypeQuestCompleted, course.NewQuestCompletedHandler(chapterSvc))
	dispatcher.Subscribe(events.EventTypeQuestCompleted, achievement.NewQuestCompletedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeChapterCompleted, concept.NewChapterCompletedHandler(loreSvc))
	dispatcher.Subscribe(events.EventTypeChapterCompleted, achievement.NewChapterCompletedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeRelicCollected, achievement.NewRelicCollectedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeDailyTurnCompleted, achievement.NewDailyTurnCompletedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeLevelReached, achievement.NewLevelReachedHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeCreativeSubmission, achievement.NewCreativeSubmissionHandler(achieveSvc))
	dispatcher.Subscribe(events.EventTypeAchievementEarned, rewardsignal.NewAchievementEarnedHandler(rewardSignalService))

	creativeSvc := gamecreative.NewCreativeServiceWithPublisher(repo.CreativeSubmissions, repo.Missions, dispatcher)
	creativeHandler := gamecreative.NewCreativeAPIHandler(creativeSvc)

	dailyTurnCfg := dailymission.DailyTurnConfig{
		XP:             config.DailyTurnXP,
		MaxTurnsPerDay: config.MaxDailyTurnsPerDay,
		Timezone:       config.Timezone,
	}
	dailyTurnSvc := dailymission.NewDailyTurnService(repo.DailyTurns, &dailyTurnCfg)
	dailyTurnSvc.SetMetrics(metrics)

	balanceSvc := balance.NewService(db.NewBalanceStore(supabaseClient))
	if err := balanceSvc.Load(ctx); err != nil {
		log.Printf("Warning: balance load failed: %v", err)
	}

	progSvc := progression.NewProgressionServiceWithPublisher(repo.Users, &progCfg, dispatcher, balanceSvc)
	progSvc.SetMetrics(metrics)

	rewardSvc := reward.NewService(repo)

	questAPIHandler := quest.NewQuestAPIHandler(questSvc, progSvc, realmStore, realmCfg, &progCfg)
	questAPIHandler.SetPublisher(dispatcher)
	questAPIHandler.SetContentGateway(contentSvc)
	questAPIHandler.SetBalance(balanceSvc)
	questAPIHandler.SetMetrics(metrics)
	questAPIHandler.SetLogger(logger)
	questAPIHandler.SetRewardService(rewardSvc)

	dailyTurnAPIHandler := dailymission.NewDailyTurnAPIHandlerWithPublisher(dailyTurnSvc, progSvc, dispatcher, balanceSvc)
	dailyTurnAPIHandler.SetMetrics(metrics)
	dailyTurnAPIHandler.SetLogger(logger)
	dailyTurnAPIHandler.SetActivityStore(repo.Activity)
	dailyTurnAPIHandler.SetRewardService(rewardSvc)

	chestEngine := chest.NewRewardEngine(repo.ChestDefinitions, repo.RelicDefinitions)
	chestSvc := chest.NewChestServiceWithPublisher(repo.Gifts, repo.PlayerRelics, repo.Collections, repo.Users, chestEngine, dispatcher)
	chestSvc.SetContentGateway(contentSvc)
	chestSvc.SetBalance(balanceSvc)
	chestSvc.SetMetrics(metrics)
	gifts.Setup(chestSvc)
	dispatcher.Subscribe(events.EventTypeQuestCompleted, chest.NewQuestCompletedHandler(chestSvc, contentSvc))

	relicSvc := relic.NewRelicService(repo.Collections, repo.PlayerRelics, repo.RelicDefinitions)
	relicSvc.SetUserStore(repo.Users)
	relicSvc.SetLedgerStore(repo.RewardLedgers)
	collections.Setup(relicSvc)

	homeSvc := gamehome.NewHomeService(questSvc, dailyTurnSvc, repo.Progression, realmStore, repo.Users, repo.CreativeSubmissions, repo.Gifts, relicSvc)
	homeSvc.SetChapterService(chapterSvc)
	homeSvc.SetLoreService(loreSvc)
	homeSvc.SetAchievementService(achieveSvc)
	homeSvc.SetCrewStreakService(familystreak.NewService(repo.Users, repo.Activity, config.Timezone))

	adminStore := db.NewDefinitionStore(supabaseClient)
	auditStore := db.NewAuditStore(supabaseClient)
	auditLogger := audit.NewLogger(auditStore)
	adminSvc := admin.NewAdminService(contentSvc, adminStore, auditLogger)
	adminSvc.SetBalance(balanceSvc)
	adminSvc.SetMetrics(metrics)
	gameAdminSvc := gameadmin.NewAdminService(supabaseClient)
	adminSvc.SetGameAdmin(gameAdminSvc)
	admin.Setup(adminSvc)

	localUserStore := db.NewLocalUserStore(supabaseClient)
	authenticator := auth.NewLocalAuthProvider(
		auth.NewBcryptHasher(),
		localUserStore,
	)

	sessionSecret := config.SessionSecret
	if sessionSecret == "" {
		sessionSecret = config.AdminSecret
	}
	issuer := auth.NewSessionIssuer(sessionSecret)
	mw := auth.NewMiddleware(issuer)

	login.Setup(authenticator, issuer, profileStore)
	me.Setup(profileStore)
	apiQuests.Setup(questAPIHandler)
	families.Setup(repo.Families, repo.Users)
	daily_missions.Setup(dailyTurnAPIHandler)

	daStore := db.NewDailyActivityEngineStore(supabaseClient)
	daSvc := dailyactivity.NewService(daStore, repo.Activity, progSvc, "Asia/Jakarta")
	daAPI := daily_activities.Setup(daSvc, logger)
	rewards.Setup(rewardSvc)
	cosmeticSvc := cosmetic.NewService(repo.Users, repo.RewardLedgers, repo.CosmeticUnlocks, profileStore, profileStore)
	apiCosmetics.Setup(cosmeticSvc)
	creative.Setup(creativeHandler)
	apiHome.Setup(homeSvc)
	homeSvc.SetSeasonService(seasonSvc)

	apiChapters.Setup(chapterSvc)
	apiLore.Setup(loreSvc)
	apiAchievements.Setup(achieveSvc)

	fragSvc := fragment.NewFragmentService(nil, repo.JourneyProgress, progSvc)
	apiStoryFragments.Setup(fragSvc)

	reactionSvc := social.NewReactionServiceWithItems(repo.Reactions, repo.CreativeSubmissions, repo.Creatives, repo.Missions)
	apiReactions.Setup(reactionSvc)

	boardSvc := board.NewService(repo.Creatives)
	apiBoard.Setup(boardSvc)

	families.Setup(repo.Families, repo.Users)
	journey_progress.Setup(repo.JourneyProgress)
	apiPush.Setup(repo.PushSubscriptions)
	apiSeasons.Setup(seasonSvc)

	// Wire Web Push delivery. VAPID keys are optional — server starts normally
	// without them, but push notifications will not be delivered.
	pushSender, pushErr := push.NewSender(push.Config{
		VAPIDPublicKey:  config.VAPIDPublicKey,
		VAPIDPrivateKey: config.VAPIDPrivateKey,
		VAPIDSubject:    config.VAPIDSubject,
	}, repo.PushSubscriptions)
	if pushErr != nil {
		log.Printf("Warning: Web Push disabled — %v", pushErr)
	} else {
		dispatcher.Subscribe(events.EventTypeDailyTurnCompleted, push.NewDailyTurnHandler(pushSender))
		dispatcher.Subscribe(events.EventTypeRelayHandoff, push.NewRelayHandoffHandler(pushSender))
	}

	secCfg := shared.DefaultSecurityConfig()
	secCfg.AllowedOrigins = config.AllowedOrigins
	secCfg.MaxBodyBytes = config.MaxBodyBytes
	secCfg.MaxBodyBytesByPath = map[string]int64{
		"/api/creative":  8 << 20,
		"/api/creative/": 8 << 20,
	}
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
			select {
			case <-bgCtx.Done():
				return
			default:
				userLimiter.Cleanup()
				loginLimiter.Cleanup()
				adminLimiter.Cleanup()
			}
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
		observability.HealthCheck{Name: "admin_store", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_journey_definitions")},
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
		// Vercel routes /api/* here, but if we catch /, just return not found or frontend.
		// For vercel, we usually don't serve the frontend from the API.
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
	mux.HandleFunc("/api/missions", secure(mw.RequireAuth(apiQuests.Handler)))
	mux.HandleFunc("/api/missions/", secure(mw.RequireAuth(apiQuests.Handler)))
	mux.HandleFunc("/api/families", secure(mw.RequireAuth(families.Handler)))
	mux.HandleFunc("/api/families/", secure(mw.RequireAuth(families.Handler)))
	mux.HandleFunc("/api/journey_progress", secure(mw.RequireAuth(journey_progress.Handler)))
	mux.HandleFunc("/api/journey_progress/", secure(mw.RequireAuth(journey_progress.Handler)))
	mux.HandleFunc("/api/daily_missions", secure(mw.RequireAuth(daily_missions.Handler)))
	mux.HandleFunc("/api/daily_missions/", secure(mw.RequireAuth(daily_missions.Handler)))
	mux.HandleFunc("/api/daily-activities/", secure(mw.RequireAuth(daAPI.Handler)))
	mux.HandleFunc("/api/creative", secure(mw.RequireAuth(csrf(creative.Handler))))
	mux.HandleFunc("/api/creative/", secure(mw.RequireAuth(csrf(creative.Handler))))
	mux.HandleFunc("/api/home", secure(mw.RequireAuth(apiHome.Handler)))
	mux.HandleFunc("/api/home/", secure(mw.RequireAuth(apiHome.Handler)))
	mux.HandleFunc("/api/gifts", secure(mw.RequireAuth(gifts.Handler)))
	mux.HandleFunc("/api/gifts/", secure(mw.RequireAuth(gifts.Handler)))
	mux.HandleFunc("/api/collections/gift", secure(mw.RequireAuth(csrf(collections.Handler))))
	mux.HandleFunc("/api/collections", secure(mw.RequireAuth(collections.Handler)))
	mux.HandleFunc("/api/collections/", secure(mw.RequireAuth(collections.Handler)))
	mux.HandleFunc("/api/courses", secure(mw.RequireAuth(apiChapters.Handler)))
	mux.HandleFunc("/api/courses/", secure(mw.RequireAuth(apiChapters.Handler)))
	mux.HandleFunc("/api/concepts", secure(mw.RequireAuth(apiLore.Handler)))
	mux.HandleFunc("/api/concepts/", secure(mw.RequireAuth(apiLore.Handler)))
	mux.HandleFunc("/api/story_fragments", secure(mw.RequireAuth(apiStoryFragments.Handler)))
	mux.HandleFunc("/api/story_fragments/", secure(mw.RequireAuth(apiStoryFragments.Handler)))
	mux.HandleFunc("/api/achievements", secure(mw.RequireAuth(apiAchievements.Handler)))
	mux.HandleFunc("/api/achievements/", secure(mw.RequireAuth(apiAchievements.Handler)))
	mux.HandleFunc("/api/reactions", secure(mw.RequireAuth(apiReactions.Handler)))
	mux.HandleFunc("/api/reactions/", secure(mw.RequireAuth(apiReactions.Handler)))
	mux.HandleFunc("/api/rewards", secure(mw.RequireAuth(rewards.Handler)))
	mux.HandleFunc("/api/rewards/", secure(mw.RequireAuth(rewards.Handler)))
	mux.HandleFunc("/api/cosmetics", secure(mw.RequireAuth(apiCosmetics.Handler)))
	mux.HandleFunc("/api/cosmetics/", secure(mw.RequireAuth(apiCosmetics.Handler)))
	mux.HandleFunc("/api/board", secure(mw.RequireAuth(csrf(apiBoard.Handler))))
	mux.HandleFunc("/api/board/", secure(mw.RequireAuth(csrf(apiBoard.Handler))))
	mux.HandleFunc("/api/push/subscribe", secure(mw.RequireAuth(csrf(apiPush.Handler))))
	mux.HandleFunc("/api/push/subscribe/", secure(mw.RequireAuth(csrf(apiPush.Handler))))
	mux.HandleFunc("/api/seasons", secure(mw.RequireAuth(apiSeasons.Handler)))
	mux.HandleFunc("/api/seasons/", secure(mw.RequireAuth(apiSeasons.Handler)))
	mux.HandleFunc("/api/admin", secure(rateLimit(adminLimiter, mw.RequireAuth(admin.Handler))))
	mux.HandleFunc("/api/admin/", secure(rateLimit(adminLimiter, mw.RequireAuth(admin.Handler))))

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

	cleanup := func(shutdownCtx context.Context) error {
		cancelBg()
		contentSvc.Stop()
		dispatcher.Close()
		userLimiter.Cleanup()
		loginLimiter.Cleanup()
		adminLimiter.Cleanup()
		logger.Close()
		return nil
	}

	return &Server{
		Handler: handler,
		Cleanup: cleanup,
	}, nil
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
