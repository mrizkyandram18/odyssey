package dailyactivity

import (
	"context"
	"errors"
)

var (
	ErrActivityNotFound = errors.New("daily activity not found")
	ErrAlreadyCompleted = errors.New("daily activity already completed today")
	ErrIncorrectAnswer  = errors.New("incorrect answer")
)

type ActivityQuestion struct {
	ID            int64    `json:"id"`
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Question      string   `json:"question"`
	Type          string   `json:"type"`
	Options       []string `json:"options"`
	CorrectAnswer string   `json:"correct_answer"`
	Explanation   string   `json:"explanation"`
	XPReward      int64    `json:"xp_reward"`
	Active        bool     `json:"active"`
}

type ActivityStore interface {
	ListActiveActivities(ctx context.Context) ([]ActivityQuestion, error)
	HasCompletedToday(ctx context.Context, uid string, date string) (bool, error)
	RecordCompletion(ctx context.Context, uid string, date string, activityID int64) error
}

type ActivityView struct {
	ID        int64    `json:"id"`
	Title     string   `json:"title"`
	Type      string   `json:"type"`
	Question  string   `json:"question"`
	Options   []string `json:"options"`
	Completed bool     `json:"completed"`
	XPReward  int64    `json:"xp_reward"`
}

type ActivityResult struct {
	Correct     bool   `json:"correct"`
	Completed   bool   `json:"completed"`
	Explanation string `json:"explanation"`
	XPAwarded   int64  `json:"xp_awarded,omitempty"`
}
