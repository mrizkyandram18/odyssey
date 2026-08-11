package creative

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"
)

// minMP4Box returns a minimal valid MP4 signature: an ftyp box whose bytes
// 4..7 == "ftyp". It passes the magic-byte check without a full playable file.
func minMP4Box() []byte {
	return []byte{0x00, 0x00, 0x00, 0x14, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0x00, 0x00, 0x00, 0x00, 'i', 's', 'o', 'm'}
}

// minWebM returns the 4-byte EBML header signature of WebM/EBML.
func minWebM() []byte {
	return []byte{0x1A, 0x45, 0xDF, 0xA3, 0x93, 0x42, 0x82, 0x88, 'm', 'a', 't', 'r', 'i', 'o', 's', 'k'}
}

// minOgg returns the OggS capture pattern.
func minOgg() []byte {
	return []byte{'O', 'g', 'g', 'S', 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02}
}

func validVideoJSON(t *testing.T, mime string, magic []byte, caption string) string {
	t.Helper()
	enc := base64.StdEncoding.EncodeToString(magic)
	return `{"v":1,"video":"data:` + mime + `;base64,` + enc + `","caption":"` + caption + `"}`
}

func TestValidateVideo_Empty(t *testing.T) {
	if err := ValidateVideo(""); err != ErrVideoEmpty {
		t.Fatalf("expected ErrVideoEmpty, got %v", err)
	}
}

func TestValidateVideo_TooLarge(t *testing.T) {
	if err := ValidateVideo(strings.Repeat("a", maxVideoBodyBytes+1)); err != ErrVideoTooLarge {
		t.Fatalf("expected ErrVideoTooLarge, got %v", err)
	}
}

func TestValidateVideo_Malformed(t *testing.T) {
	if err := ValidateVideo("not json"); err != ErrVideoMalformed {
		t.Fatalf("expected ErrVideoMalformed, got %v", err)
	}
}

func TestValidateVideo_MissingVideo(t *testing.T) {
	if err := ValidateVideo(`{"v":1,"caption":"hi"}`); err != ErrVideoMissing {
		t.Fatalf("expected ErrVideoMissing, got %v", err)
	}
}

func TestValidateVideo_BadURI(t *testing.T) {
	if err := ValidateVideo(`{"v":1,"video":"http://example.com/clip.mp4"}`); err != ErrVideoBadURI {
		t.Fatalf("expected ErrVideoBadURI, got %v", err)
	}
}

func TestValidateVideo_NotVideo(t *testing.T) {
	enc := base64.StdEncoding.EncodeToString([]byte("plain text"))
	if err := ValidateVideo(`{"v":1,"video":"data:text/plain;base64,` + enc + `"}`); err != ErrVideoNotVideo {
		t.Fatalf("expected ErrVideoNotVideo, got %v", err)
	}
}

func TestValidateVideo_InvalidBase64(t *testing.T) {
	if err := ValidateVideo(`{"v":1,"video":"data:video/mp4;base64,!!!notbase64!!!"}`); err != ErrVideoDecode {
		t.Fatalf("expected ErrVideoDecode, got %v", err)
	}
}

func TestValidateVideo_NoBase64(t *testing.T) {
	if err := ValidateVideo(`{"v":1,"video":"data:video/mp4,rawbytes"}`); err != ErrVideoBadURI {
		t.Fatalf("expected ErrVideoBadURI, got %v", err)
	}
}

// Mismatched MIME and container signature: declared video/mp4 but payload is
// plain text with no ftyp box. Must be rejected by the magic-byte check.
func TestValidateVideo_BadMagicMP4(t *testing.T) {
	enc := base64.StdEncoding.EncodeToString([]byte("not an mp4 at all"))
	if err := ValidateVideo(`{"v":1,"video":"data:video/mp4;base64,` + enc + `"}`); err != ErrVideoBadMagic {
		t.Fatalf("expected ErrVideoBadMagic, got %v", err)
	}
}

// Unsupported video MIME (video/quicktime) is not in the accepted set, so even
// an ftyp-bearing payload is rejected.
func TestValidateVideo_UnsupportedMIME(t *testing.T) {
	enc := base64.StdEncoding.EncodeToString(minMP4Box())
	if err := ValidateVideo(`{"v":1,"video":"data:video/quicktime;base64,` + enc + `"}`); err != ErrVideoBadMagic {
		t.Fatalf("expected ErrVideoBadMagic, got %v", err)
	}
}

func TestValidateVideo_CaptionTooLong(t *testing.T) {
	long := strings.Repeat("a", maxVideoCaption+1)
	if err := ValidateVideo(validVideoJSON(t, "video/mp4", minMP4Box(), long)); err != ErrVideoCaptionLong {
		t.Fatalf("expected ErrVideoCaptionLong, got %v", err)
	}
}

func TestValidateVideo_TooBigDecoded(t *testing.T) {
	// bytes longer than maxVideoBytes but carrying a valid ftyp signature so the
	// magic check passes and the size check is what trips.
	big := make([]byte, maxVideoBytes+1)
	copy(big[4:8], []byte("ftyp"))
	enc := base64.StdEncoding.EncodeToString(big)
	if err := ValidateVideo(`{"v":1,"video":"data:video/mp4;base64,` + enc + `"}`); err != ErrVideoTooBig {
		t.Fatalf("expected ErrVideoTooBig, got %v", err)
	}
}

func TestValidateVideo_SuccessMP4(t *testing.T) {
	if err := ValidateVideo(validVideoJSON(t, "video/mp4", minMP4Box(), "a sunny clip")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateVideo_SuccessWebM(t *testing.T) {
	if err := ValidateVideo(validVideoJSON(t, "video/webm", minWebM(), "w")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateVideo_SuccessOgg(t *testing.T) {
	if err := ValidateVideo(validVideoJSON(t, "video/ogg", minOgg(), "w")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateVideo_SuccessNoCaption(t *testing.T) {
	enc := base64.StdEncoding.EncodeToString(minMP4Box())
	if err := ValidateVideo(`{"v":1,"video":"data:video/mp4;base64,` + enc + `"}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// errors.Is should resolve the sentinel validators.
func TestValidateVideo_ErrorSentinels(t *testing.T) {
	if !errors.Is(ValidateVideo(""), ErrVideoEmpty) {
		t.Fatal("sentinel mismatch")
	}
}
