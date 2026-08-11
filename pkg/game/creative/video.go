package creative

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"mime"
	"strings"
)

var (
	ErrVideoEmpty       = errors.New("video payload is empty")
	ErrVideoTooLarge    = errors.New("video payload exceeds maximum size")
	ErrVideoMalformed   = errors.New("video payload is malformed json")
	ErrVideoMissing     = errors.New("video payload missing 'video' data uri")
	ErrVideoBadURI      = errors.New("video data uri is malformed")
	ErrVideoNotVideo    = errors.New("video data uri must be a video")
	ErrVideoDecode      = errors.New("video data uri is not valid base64")
	ErrVideoTooBig      = errors.New("video exceeds maximum decoded size")
	ErrVideoCaptionLong = errors.New("video caption is too long")
	ErrVideoBadMagic    = errors.New("video data uri has unsupported container signature")
)

const (
	// maxVideoBodyBytes bounds the raw JSON payload length for a VIDEO submission.
	// This matches the 8 MiB per-route HTTP body cap enforced on /api/creative
	// (see pkg/server/server.go and pkg/shared/security.go).
	maxVideoBodyBytes = 8 * 1024 * 1024
	// maxVideoBytes bounds the decoded video bytes. A base64 video is ~1.33x its
	// raw size, so 5 MiB decoded (~6.7 MiB base64) stays comfortably under the
	// 8 MiB body cap plus JSON envelope, leaving ErrVideoTooBig reachable as
	// defense-in-depth rather than dead code.
	//
	// NOTE: this inline-base64 architecture only supports very short, low-quality
	// clips (a few seconds). Realistic smartphone video is many tens of MiB and
	// is NOT supported without a dedicated upload/storage pipeline (see
	// docs/roadmap.md Phase 2 blockers: VIDEO storage). Do not claim this is
	// production-ready for long-form video.
	maxVideoBytes   = 5 * 1024 * 1024
	maxVideoCaption = 280
)

// VideoPayload is the structured content for SubmissionVideo.
// Stored as a JSON string in Submission.Content (TEXT column).
type VideoPayload struct {
	// V is an optional schema version (defaults to 1).
	V       int    `json:"v,omitempty"`
	Video   string `json:"video"`
	Caption string `json:"caption,omitempty"`
}

// ValidateVideo validates a VIDEO submission content string.
// Format: {"v":1,"video":"data:video/mp4;base64,...","caption":"..."}
// Rules:
//   - non-empty, max maxVideoBodyBytes runes
//   - well-formed JSON with a "video" data URI
//   - video URI must be video/* and ;base64
//   - base64 must decode and decoded bytes <= maxVideoBytes
//   - decoded bytes must carry a supported container signature (magic bytes)
//   - caption <= 280 runes (optional)
func ValidateVideo(payload string) error {
	if payload == "" {
		return ErrVideoEmpty
	}
	if len(payload) > maxVideoBodyBytes {
		return ErrVideoTooLarge
	}

	var vp VideoPayload
	if err := json.Unmarshal([]byte(payload), &vp); err != nil {
		return ErrVideoMalformed
	}

	dataURI := strings.TrimSpace(vp.Video)
	if dataURI == "" {
		return ErrVideoMissing
	}

	videoBytes, err := parseVideoURL(dataURI)
	if err != nil {
		return err
	}
	if len(videoBytes) > maxVideoBytes {
		return ErrVideoTooBig
	}

	if utfCount(vp.Caption) > maxVideoCaption {
		return ErrVideoCaptionLong
	}

	return nil
}

// supportedVideoMagics maps an accepted video MIME type to a checker for its
// container signature. The data URI's declared media type is trusted only after
// this byte-level signature check, so a mislabeled or corrupt payload is
// rejected (ErrVideoBadMagic) even if the declared MIME is video/*.
var supportedVideoMagics = map[string]func([]byte) bool{
	"video/mp4":  hasMP4Magic,
	"video/webm": hasWebMMagic,
	"video/ogg":  hasOggMagic,
}

// parseVideoURL extracts and validates a base64-encoded video data URI.
// Requires scheme "data:" with a media type and the ";base64" token, the media
// type must be video/*, the base64 must decode, and the decoded bytes must
// carry a supported container signature.
func parseVideoURL(uri string) ([]byte, error) {
	const prefix = "data:"
	if !strings.HasPrefix(uri, prefix) {
		return nil, ErrVideoBadURI
	}
	rest := uri[len(prefix):]

	comma := strings.IndexByte(rest, ',')
	if comma < 0 {
		return nil, ErrVideoBadURI
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
		return nil, ErrVideoBadURI
	}

	mediaType = strings.TrimSpace(mediaType)
	if mediaType == "" {
		mediaType = "text/plain"
	}
	mt, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return nil, ErrVideoBadURI
	}
	if !strings.HasPrefix(mt, "video/") {
		return nil, ErrVideoNotVideo
	}

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil, ErrVideoDecode
	}

	check, ok := supportedVideoMagics[mt]
	if !ok || !check(decoded) {
		return nil, ErrVideoBadMagic
	}

	return decoded, nil
}

// hasMP4Magic reports whether b begins with an ISO base media "ftyp" box
// (bytes 4..7 == "ftyp"), the signature of MP4/MOV files.
func hasMP4Magic(b []byte) bool {
	return len(b) >= 8 && string(b[4:8]) == "ftyp"
}

// hasWebMMagic reports whether b begins with the EBML header signature used by
// WebM/EBML files (1A 45 DF A3).
func hasWebMMagic(b []byte) bool {
	return len(b) >= 4 && b[0] == 0x1A && b[1] == 0x45 && b[2] == 0xDF && b[3] == 0xA3
}

// hasOggMagic reports whether b begins with the Ogg capture pattern "OggS"
// (4F 67 67 53), the signature of Ogg-encapsulated video.
func hasOggMagic(b []byte) bool {
	return len(b) >= 4 && string(b[0:4]) == "OggS"
}
