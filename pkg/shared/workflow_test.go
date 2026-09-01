package shared

import (
	"os"
	"testing"
)

func TestNoAutoBlockWorkflow(t *testing.T) {
	path := "../../.github/workflows/auto-block-inactive-users.yml"
	_, err := os.ReadFile(path)
	if err == nil {
		t.Fatalf("workflow file should have been removed, but still exists at %s", path)
	}
	// Ensure no executable auto-block code remains (grep handled via this test's absence)
	// Documentation may still mention migration history, but no scheduler file should exist
}
