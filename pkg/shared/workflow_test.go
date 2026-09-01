package shared

import (
	"os"
	"strings"
	"testing"
)

func TestAutoBlockWorkflowValidation(t *testing.T) {
	path := "../../.github/workflows/auto-block-inactive-users.yml"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("workflow file not found at %s: %v", path, err)
	}
	content := string(data)
	// cron exactly 5 17 * * *
	if !strings.Contains(content, "cron: '5 17 * * *'") {
		t.Fatalf("expected cron '5 17 * * *' not found")
	}
	// workflow_dispatch exists
	if !strings.Contains(content, "workflow_dispatch:") {
		t.Fatalf("workflow_dispatch not found")
	}
	// POST
	if !strings.Contains(content, "-X POST") {
		t.Fatalf("expected POST method not found")
	}
	// fail on non-2xx
	if !strings.Contains(content, "HTTP_CODE") || !strings.Contains(content, "ge 300") {
		t.Fatalf("expected fail on non-2xx handling")
	}
	// timeout
	if !strings.Contains(content, "--max-time 30") {
		t.Fatalf("expected timeout --max-time 30")
	}
	// concurrency
	if !strings.Contains(content, "concurrency:") {
		t.Fatalf("expected concurrency protection")
	}
	// no plaintext secret/token hardcoded (should use secrets.*)
	if strings.Contains(content, "ODYSSEY_AUTO_BLOCK_TOKEN=") && strings.Contains(content, "dev-") {
		t.Fatalf("plaintext token found")
	}
	// Ensure secrets are referenced, not hardcoded values
	if !strings.Contains(content, "secrets.ODYSSEY_PRODUCTION_URL") {
		t.Fatalf("expected ODYSSEY_PRODUCTION_URL secret reference")
	}
	if !strings.Contains(content, "secrets.ODYSSEY_AUTO_BLOCK_TOKEN") {
		t.Fatalf("expected ODYSSEY_AUTO_BLOCK_TOKEN secret reference")
	}
	// endpoint
	if !strings.Contains(content, "/api/admin/members/auto-block") {
		t.Fatalf("expected endpoint /api/admin/members/auto-block")
	}
	// No hardcoded production hostname
	if strings.Contains(content, "odyssey.vercel.app") || strings.Contains(content, "example.supabase.co") {
		t.Fatalf("hardcoded production hostname found")
	}
	// Check no JWT or service-role key plaintext
	lower := strings.ToLower(content)
	if strings.Contains(lower, "jwt") && strings.Contains(lower, "eyj") {
		t.Fatalf("plaintext JWT found")
	}
}
