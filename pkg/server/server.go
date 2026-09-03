package server

import (
	"context"
	"net/http"
	"time"

	apiAdminMembers "odyssey/internal/api/admin_members"
	apiAdminTasks "odyssey/internal/api/admin_tasks"
	"odyssey/internal/api/families"
	apiFamilyTasks "odyssey/internal/api/family_tasks"
	"odyssey/internal/api/login"
	"odyssey/internal/api/me"
	apiPayoutConfig "odyssey/internal/api/payout_config"
	"odyssey/internal/api/push"
	apiShop "odyssey/internal/api/shop"
	"odyssey/internal/api/status"
	"odyssey/pkg/auth"
	"odyssey/pkg/db"
	"odyssey/pkg/observability"
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
	localUserStore := db.NewLocalUserStore(supabaseClient)
	familyStore := db.NewFamilyStore(supabaseClient)
	pushStore := db.NewPushSubscriptionStore(supabaseClient)

	metrics := observability.NewMetrics()
	profiler := observability.NewProfiler()
	logger := observability.DefaultLogger()

	schemaVer, readErr := observability.ReadSchemaVersion(ctx, supabaseClient)
	if readErr != nil {
		logger.Warn("schema_version_read_failed", map[string]any{
			"error": readErr.Error(),
		})
	} else {
		observability.SetRuntimeSchemaVersion(schemaVer)
		logger.Info("schema_version_loaded", map[string]any{
			"schema_version": schemaVer,
		})
	}

	bgCtx, cancelBg := context.WithCancel(context.Background())

	authenticator := auth.NewLocalAuthProviderWithBinder(
		auth.NewBcryptHasher(),
		localUserStore,
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
	families.Setup(familyStore, nil)
	push.Setup(pushStore)

	familyTasksAPI := apiFamilyTasks.NewAPI(supabaseClient)
	adminTasksAPI := apiAdminTasks.NewAPI(supabaseClient)
	adminMembersAPI := apiAdminMembers.NewAPI(supabaseClient)
	shopAPI := apiShop.NewAPI(supabaseClient)
	payoutConfigAPI := apiPayoutConfig.NewAPI(supabaseClient)

	secCfg := shared.DefaultSecurityConfig()
	secCfg.AllowedOrigins = config.AllowedOrigins
	secCfg.MaxBodyBytes = config.MaxBodyBytes
	secCfg.MaxBodyBytesByPath = map[string]int64{
		"/api/tasks/upload": 10 << 20,
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
		for {
			select {
			case <-bgCtx.Done():
				return
			case <-ticker.C:
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
		observability.HealthCheck{Name: "database", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_user_profiles")},
		observability.HealthCheck{Name: "tasks_store", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_tasks")},
		observability.HealthCheck{Name: "coin_ledger", Fn: observability.DBHealthCheck(supabaseClient, "odyssey_coin_transactions")},
	)

	status.Setup(status.FuncStatusProvider(func(ctx context.Context) map[string]any {
		return map[string]any{
			"app":            "odyssey",
			"version":        observability.Version,
			"schema_version": observability.SchemaVersionString(),
			"uptime_seconds": int64(time.Since(metrics.BootTime()).Seconds()),
			"boot_time":      metrics.BootTime().UTC().Format(time.RFC3339),
		}
	}))

	mux := http.NewServeMux()

	secure := func(next http.HandlerFunc) http.HandlerFunc {
		return shared.SecurityHeadersMiddleware(secCfg, shared.CORSHeaderMiddleware(secCfg, shared.RequestLimitMiddleware(secCfg, next)))
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

	// Auth & Identity
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

	// Family & Push Notifications
	mux.HandleFunc("/api/families", secure(mw.RequireAuth(families.Handler)))
	mux.HandleFunc("/api/families/", secure(mw.RequireAuth(families.Handler)))
	mux.HandleFunc("/api/push", secure(mw.RequireAuth(push.Handler)))
	mux.HandleFunc("/api/push/", secure(mw.RequireAuth(push.Handler)))

	// Core Family Daily Tasks
	mux.HandleFunc("/api/tasks", secure(mw.RequireAuth(familyTasksAPI.Handler)))
	mux.HandleFunc("/api/tasks/", secure(mw.RequireAuth(familyTasksAPI.Handler)))

	// Core Reward Shop & Claims
	mux.HandleFunc("/api/shop", secure(mw.RequireAuth(shopAPI.Handler)))
	mux.HandleFunc("/api/shop/", secure(mw.RequireAuth(shopAPI.Handler)))

	// Admin Panel: Members, Tasks, Verification Queue, Claims Payout
	mux.HandleFunc("/api/admin/members", secure(rateLimit(adminLimiter, mw.RequireAuth(adminMembersAPI.Handler))))
	mux.HandleFunc("/api/admin/members/", secure(rateLimit(adminLimiter, mw.RequireAuth(adminMembersAPI.Handler))))
	mux.HandleFunc("/api/admin/tasks", secure(rateLimit(adminLimiter, mw.RequireAuth(adminTasksAPI.Handler))))
	mux.HandleFunc("/api/admin/tasks/", secure(rateLimit(adminLimiter, mw.RequireAuth(adminTasksAPI.Handler))))
	mux.HandleFunc("/api/admin/submissions", secure(rateLimit(adminLimiter, mw.RequireAuth(adminTasksAPI.Handler))))
	mux.HandleFunc("/api/admin/submissions/", secure(rateLimit(adminLimiter, mw.RequireAuth(adminTasksAPI.Handler))))
	mux.HandleFunc("/api/admin/claims", secure(rateLimit(adminLimiter, mw.RequireAuth(shopAPI.Handler))))
	mux.HandleFunc("/api/admin/claims/", secure(rateLimit(adminLimiter, mw.RequireAuth(shopAPI.Handler))))
	mux.HandleFunc("/api/admin/config", secure(rateLimit(adminLimiter, mw.RequireAuth(adminTasksAPI.Handler))))
	mux.HandleFunc("/api/admin/config/", secure(rateLimit(adminLimiter, mw.RequireAuth(adminTasksAPI.Handler))))
	mux.HandleFunc("/api/admin/payout-config", secure(rateLimit(adminLimiter, mw.RequireAuth(payoutConfigAPI.Handler))))
	mux.HandleFunc("/api/admin/payout-config/", secure(rateLimit(adminLimiter, mw.RequireAuth(payoutConfigAPI.Handler))))

	// Observability & Health
	mux.HandleFunc("/metrics", secure(observability.InternalTokenMiddleware(config.InternalMetricsToken, observability.MetricsHandler(metrics))))
	mux.HandleFunc("/version", secure(observability.VersionHandler(nil)))
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
