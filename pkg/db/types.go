package db

import (
	"time"
)

type UserProfile struct {
	UID          string    `json:"uid"`
	FamilyID     string    `json:"family_id"`
	ExplorerName string    `json:"explorer_name"`
	Role         string    `json:"role"`
	Level        int       `json:"level"`
	XP           int64     `json:"xp"`
	Coins        int64     `json:"coins"`
	StreakDays   int       `json:"streak_days"`
	LastActive   *string   `json:"last_active_date,omitempty"`
	AvatarStyle  string    `json:"avatar_style"`
	AvatarSeed   string    `json:"avatar_seed"`
	Version      int       `json:"version"`
	PasswordHash string    `json:"-"`
	DeviceID           string    `json:"device_id,omitempty"`
	IsActive           bool      `json:"is_active"`
	MustChangePassword bool      `json:"must_change_password"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Family struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	BannerURL string    `json:"banner_url,omitempty"`
	Theme     string    `json:"theme,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PushSubscription struct {
	ID        int64     `json:"id"`
	UID       string    `json:"uid"`
	Endpoint  string    `json:"endpoint"`
	P256DH    string    `json:"p256dh"`
	Auth      string    `json:"auth"`
	CreatedAt time.Time `json:"created_at"`
}
