package dailymission

import (
	"errors"
	"time"
)

var ErrDailyTurnNotFound = errors.New("daily turn not found")
var ErrNoTurnsRemaining = errors.New("no turns remaining")

type DailyTurnStatus string

const (
	DailyTurnStatusPending DailyTurnStatus = "PENDING"
	DailyTurnStatusDone    DailyTurnStatus = "DONE"
)

// DailyTurnXP is the default XP awarded per daily turn.
// Deprecated: use DailyTurnConfig.XP instead.
const DailyTurnXP int64 = 30

// DailyTurnConfig carries configurable daily turn parameters.
type DailyTurnConfig struct {
	// XP is the amount of XP awarded on turn consumption (default: 30).
	XP int64
	// MaxTurnsPerDay is the maximum number of turns a player may consume
	// per calendar day (default: 1). Values < 1 are treated as 1.
	MaxTurnsPerDay int
	// Timezone is the IANA timezone used to compute the calendar date for
	// turn availability checks (default: "Asia/Jakarta").
	Timezone string
	// Now is an injectable clock; defaults to time.Now for production use.
	// Tests supply a fixed time to make date math deterministic.
	Now func() time.Time
}

// DefaultDailyTurnConfig returns the standard MVP daily turn values.
func DefaultDailyTurnConfig() DailyTurnConfig {
	return DailyTurnConfig{
		XP:             30,
		MaxTurnsPerDay: 1,
		Timezone:       "UTC",
		Now:            time.Now,
	}
}
