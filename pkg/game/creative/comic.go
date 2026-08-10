package creative

import (
	"encoding/json"
	"errors"
	"strings"
	"unicode/utf8"
)

var (
	ErrComicEmpty         = errors.New("comic payload is empty")
	ErrComicTooLarge      = errors.New("comic payload exceeds maximum size")
	ErrComicMalformed     = errors.New("comic payload is malformed json")
	ErrComicPanelCount    = errors.New("comic must have between 2 and 4 panels")
	ErrComicPanelEmpty    = errors.New("each comic panel needs a caption or a sketch")
	ErrComicCaptionTooLong = errors.New("comic panel caption is too long")
)

const (
	// maxComicBodyBytes keeps multi-panel comics under the default 1 MiB HTTP body limit.
	maxComicBodyBytes = 750 * 1024
	minComicPanels    = 2
	maxComicPanels    = 4
	maxComicCaption   = 280
)

// ComicPanel is one strip panel: caption text and/or an optional SVG sketch.
type ComicPanel struct {
	Caption string `json:"caption"`
	SVG     string `json:"svg,omitempty"`
}

// ComicPayload is the structured content for SubmissionComic.
// Stored as a JSON string in Submission.Content (TEXT column).
type ComicPayload struct {
	// V is an optional schema version (defaults to 1).
	V      int          `json:"v,omitempty"`
	Panels []ComicPanel `json:"panels"`
}

// ValidateComic validates a COMIC submission content string.
// Format: {"v":1,"panels":[{"caption":"...","svg":"<svg>...</svg>"}]}
// Rules:
//   - non-empty, max 750 KiB
//   - well-formed JSON with 2–4 panels
//   - each panel has a non-empty caption and/or SVG
//   - caption ≤ 280 runes
//   - any SVG is checked with ValidateSVG
func ValidateComic(payload string) error {
	if payload == "" {
		return ErrComicEmpty
	}
	if len(payload) > maxComicBodyBytes {
		return ErrComicTooLarge
	}

	var comic ComicPayload
	if err := json.Unmarshal([]byte(payload), &comic); err != nil {
		return ErrComicMalformed
	}

	n := len(comic.Panels)
	if n < minComicPanels || n > maxComicPanels {
		return ErrComicPanelCount
	}

	for _, p := range comic.Panels {
		caption := strings.TrimSpace(p.Caption)
		svg := strings.TrimSpace(p.SVG)
		if caption == "" && svg == "" {
			return ErrComicPanelEmpty
		}
		if utf8.RuneCountInString(caption) > maxComicCaption {
			return ErrComicCaptionTooLong
		}
		if svg != "" {
			if err := ValidateSVG(svg); err != nil {
				return err
			}
		}
	}

	return nil
}
