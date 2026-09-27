package db

import (
	"testing"
)

func TestValidateTable(t *testing.T) {
	tables := []string{
		"odyssey_user_profiles",
		"odyssey_families",
		"odyssey_tasks",
		"odyssey_task_submissions",
		"odyssey_reward_catalog",
		"odyssey_claims",
		"odyssey_coin_transactions",
		"odyssey_push_subscriptions",
		"odyssey_schema_version",
		"odyssey_system_config",
		"odyssey_user_payout_config",
		"odyssey_member_monthly_targets",
		"odyssey_historical_compensation_grants",
		"reward_tickets",
		"user_cosmetics",
		"cosmetic_items",
	}

	for _, tbl := range tables {
		if err := validateTable(tbl); err != nil {
			t.Errorf("expected table %s to be allowed, got: %v", tbl, err)
		}
	}

	disallowed := []string{
		"users",
		"admin_passwords",
		"pg_shadow",
		"",
	}
	for _, tbl := range disallowed {
		if err := validateTable(tbl); err == nil {
			t.Errorf("expected disallowed table '%s' to fail validation", tbl)
		}
	}
}
