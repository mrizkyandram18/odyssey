package fragment

import (
	"time"

	"odyssey/pkg/game"
)

// StoryFragmentView is the client-facing view of a story fragment in the Journal.
type StoryFragmentView struct {
	Slug         string     `json:"slug"`
	Journey        string     `json:"journey"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	SetName      string     `json:"set_name"`
	IsHidden     bool       `json:"is_hidden"`
	Discovered   bool       `json:"discovered"`
	DiscoveredAt *time.Time `json:"discovered_at,omitempty"`
}

// DiscoverResult is returned when a player discovers/collects a fragment.
type DiscoverResult struct {
	Fragment   StoryFragmentView `json:"fragment"`
	Discovered bool              `json:"discovered"` // true if newly discovered
	XPGranted  int64             `json:"xp_granted"`
}

// ReplayResult is returned when a player replays a completed journey.
type ReplayResult struct {
	Journey             string              `json:"journey"`
	IsReplay          bool                `json:"is_replay"`
	BonusDialogue     string              `json:"bonus_dialogue"`
	UnlockedFragments []StoryFragmentView `json:"unlocked_fragments"`
}

// DefaultFragments is the embedded seed catalog used when no external store is loaded.
var DefaultFragments = []game.LearningConcept{
	{
		Slug:     "ancient-bark-whisper",
		Journey:    "whispering-woods",
		Title:    "Bisikan Pepohonan Tua",
		Content:  "Pohon-pohon raksasa di Hutan Berbisik menyimpan gema langkah penjelajah pertama.",
		SetName:  "whispering-set",
		IsHidden: false,
	},
	{
		Slug:     "echo-of-the-first-explorer",
		Journey:    "whispering-woods",
		Title:    "Gema Penjelajah Perdana",
		Content:  "Rahasia Replay: Di balik lumut tua, terukir ukiran kompas kuno yang ditinggalkan ribuan purnama lalu.",
		SetName:  "whispering-set",
		IsHidden: true,
	},
	{
		Slug:     "copper-cog-diagram",
		Journey:    "clockwork-city",
		Title:    "Bagan Roda Gigi Tembaga",
		Content:  "Diagram kuno yang menunjukkan susunan roda gigi raksasa di pusat Kota Jam.",
		SetName:  "clockwork-set",
		IsHidden: false,
	},
	{
		Slug:     "secret-steam-valve",
		Journey:    "clockwork-city",
		Title:    "Katup Uap Rahasia",
		Content:  "Rahasia Replay: Katup kuningan yang menyembunyikan lorong ruang uap tak tersentuh.",
		SetName:  "clockwork-set",
		IsHidden: true,
	},
}
