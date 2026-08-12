package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
)

func TestAdminEndpointsAuth(t *testing.T) {
	// Setup dummy service to avoid 503
	Setup(&AdminService{})

	newP2Endpoints := []string{
		"/api/admin/stats",
		"/api/admin/missions",
		"/api/admin/daily-activities",
		"/api/admin/missions/test/toggle",
		"/api/admin/daily-activities/1/toggle",
	}

	oldEndpoints := []string{
		"/api/admin/reload",
		"/api/admin/status",
		"/api/admin/metrics",
		"/api/admin/validate",
		"/api/admin/audit",
		"/api/admin/balance",
	}

	testCases := []struct {
		Role          auth.Role
		ExpectedP2    bool // true means allowed (not 401/403)
		ExpectedOld   bool // true means allowed (not 401/403)
	}{
		{Role: auth.RoleSeeker, ExpectedP2: false, ExpectedOld: false},
		{Role: auth.RoleGuide, ExpectedP2: true, ExpectedOld: false},
		{Role: auth.RoleAdmin, ExpectedP2: true, ExpectedOld: true},
	}

	runEndpoints := func(endpoints []string, role auth.Role, shouldAllow bool) {
		for _, ep := range endpoints {
			name := string(role) + "_" + ep
			t.Run(name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, ep, nil)
				if ep[len(ep)-6:] == "toggle" || ep == "/api/admin/reload" {
					req = httptest.NewRequest(http.MethodPost, ep, nil)
				}
				
				ctx := auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
					Role: string(role),
				})
				req = req.WithContext(ctx)

				rr := httptest.NewRecorder()
				func() {
					defer func() { recover() }()
					Handler(rr, req)
				}()

				allowed := rr.Code != http.StatusUnauthorized && rr.Code != http.StatusForbidden

				if allowed != shouldAllow {
					t.Errorf("endpoint %s for role %s: expected allowed=%v, got allowed=%v (code %d, body: %s)", ep, role, shouldAllow, allowed, rr.Code, rr.Body.String())
				}
			})
		}
	}

	for _, tc := range testCases {
		runEndpoints(newP2Endpoints, tc.Role, tc.ExpectedP2)
		runEndpoints(oldEndpoints, tc.Role, tc.ExpectedOld)
	}
}
