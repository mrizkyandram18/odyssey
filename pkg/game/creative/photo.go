package creative

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"mime"
	"strings"
)

var (
	ErrPhotoEmpty       = errors.New("photo payload is empty")
	ErrPhotoTooLarge    = errors.New("photo payload exceeds maximum size")
	ErrPhotoMalformed   = errors.New("photo payload is malformed json")
	ErrPhotoMissing     = errors.New("photo payload missing 'photo' data uri")
	ErrPhotoBadURI      = errors.New("photo data uri is malformed")
	ErrPhotoNotImage    = errors.New("photo data uri must be an image")
	ErrPhotoDecode      = errors.New("photo data uri is not valid base64")
	ErrPhotoTooBig      = errors.New("photo image exceeds maximum decoded size")
	ErrPhotoCaptionLong = errors.New("photo caption is too long")
)

const (
	// maxPhotoBodyBytes bounds the raw JSON payload length for a PHOTO submission.
	// This matches the 8 MiB per-route HTTP body cap enforced on /api/creative.
	maxPhotoBodyBytes = 8 * 1024 * 1024
	// maxPhotoBytes bounds the decoded image bytes. A base64 image is ~1.33x its
	// raw size, so 5 MiB decoded (~6.7 MiB base64) stays comfortably under the
	// 8 MiB body cap plus JSON envelope, leaving ErrPhotoTooBig reachable as
	// defense-in-depth rather than dead code.
	maxPhotoBytes   = 5 * 1024 * 1024
	maxPhotoCaption = 280
)

// PhotoPayload is the structured content for SubmissionPhoto.
// Stored as a JSON string in Submission.Content (TEXT column).
type PhotoPayload struct {
	// V is an optional schema version (defaults to 1).
	V       int    `json:"v,omitempty"`
	Photo   string `json:"photo"`
	Caption string `json:"caption,omitempty"`
}

// ValidatePhoto validates a PHOTO submission content string.
// Format: {"v":1,"photo":"data:image/jpeg;base64,...","caption":"..."}
// Rules:
//   - non-empty, max maxPhotoBodyBytes runes
//   - well-formed JSON with a photo data URI
//   - photo URI must be image/* and ;base64
//   - base64 must decode and decoded bytes <= maxPhotoBytes
//   - caption <= 280 runes (optional)
func ValidatePhoto(payload string) error {
	if payload == "" {
		return ErrPhotoEmpty
	}
	if len(payload) > maxPhotoBodyBytes {
		return ErrPhotoTooLarge
	}

	var photo PhotoPayload
	if err := json.Unmarshal([]byte(payload), &photo); err != nil {
		return ErrPhotoMalformed
	}

	dataURI := strings.TrimSpace(photo.Photo)
	if dataURI == "" {
		return ErrPhotoMissing
	}

	photoBytes, err := parseDataURL(dataURI)
	if err != nil {
		return err
	}
	if len(photoBytes) > maxPhotoBytes {
		return ErrPhotoTooBig
	}

	if utfCount(photo.Caption) > maxPhotoCaption {
		return ErrPhotoCaptionLong
	}

	return nil
}

// parseDataURL extracts and validates a base64-encoded data URI.
// Requires scheme "data:" with a media type and the ";base64" token,
// and the media type must be an image/*.
func parseDataURL(uri string) ([]byte, error) {
	const prefix = "data:"
	if !strings.HasPrefix(uri, prefix) {
		return nil, ErrPhotoBadURI
	}
	rest := uri[len(prefix):]

	comma := strings.IndexByte(rest, ',')
	if comma < 0 {
		return nil, ErrPhotoBadURI
	}

	meta := rest[:comma]
	data := rest[comma+1:]

	isBase64 := false
	mediaType := meta
	if i := strings.LastIndex(meta, ";base64"); i >= 0 {
		isBase64 = true
		mediaType = meta[:i]
	}
	if !isBase64 {
		return nil, ErrPhotoBadURI
	}

	mediaType = strings.TrimSpace(mediaType)
	if mediaType == "" {
		mediaType = "text/plain"
	}
	// A data URI may carry a charset token we don't care about; strip parameters
	// after the first ';' so mime.ParseMediaType sees a clean type+params.
	mt, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, ErrPhotoBadURI
	}
	if !strings.HasPrefix(mt, "image/") {
		return nil, ErrPhotoNotImage
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, ErrPhotoDecode
	}
	return decoded, nil
}

func utfCount(s string) int {
	count := 0
	for range s {
		count++
	}
	return count
}
