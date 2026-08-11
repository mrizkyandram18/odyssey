package creative

import (
	"encoding/base64"
	"strings"
	"testing"
)

func validPhotoJSON(t *testing.T, caption string) string {
	t.Helper()
	png := []byte{0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a}
	return `{"v":1,"photo":"data:image/png;base64,` + base64.StdEncoding.EncodeToString(png) + `","caption":"` + caption + `"}`
}

func TestValidatePhoto_Empty(t *testing.T) {
	if err := ValidatePhoto(""); err != ErrPhotoEmpty {
		t.Fatalf("expected ErrPhotoEmpty, got %v", err)
	}
}

func TestValidatePhoto_TooLarge(t *testing.T) {
	if err := ValidatePhoto(strings.Repeat("a", maxPhotoBodyBytes+1)); err != ErrPhotoTooLarge {
		t.Fatalf("expected ErrPhotoTooLarge, got %v", err)
	}
}

func TestValidatePhoto_Malformed(t *testing.T) {
	if err := ValidatePhoto("not json"); err != ErrPhotoMalformed {
		t.Fatalf("expected ErrPhotoMalformed, got %v", err)
	}
}

func TestValidatePhoto_MissingPhoto(t *testing.T) {
	if err := ValidatePhoto(`{"v":1,"caption":"hi"}`); err != ErrPhotoMissing {
		t.Fatalf("expected ErrPhotoMissing, got %v", err)
	}
}

func TestValidatePhoto_BadURI(t *testing.T) {
	if err := ValidatePhoto(`{"v":1,"photo":"http://example.com/photo.png"}`); err != ErrPhotoBadURI {
		t.Fatalf("expected ErrPhotoBadURI, got %v", err)
	}
}

func TestValidatePhoto_NotImage(t *testing.T) {
	enc := base64.StdEncoding.EncodeToString([]byte("plain text"))
	if err := ValidatePhoto(`{"v":1,"photo":"data:text/plain;base64,` + enc + `"}`); err != ErrPhotoNotImage {
		t.Fatalf("expected ErrPhotoNotImage, got %v", err)
	}
}

func TestValidatePhoto_InvalidBase64(t *testing.T) {
	if err := ValidatePhoto(`{"v":1,"photo":"data:image/png;base64,!!!notbase64!!!"}`); err != ErrPhotoDecode {
		t.Fatalf("expected ErrPhotoDecode, got %v", err)
	}
}

func TestValidatePhoto_NoBase64(t *testing.T) {
	if err := ValidatePhoto(`{"v":1,"photo":"data:image/png,rawbytes"}`); err != ErrPhotoBadURI {
		t.Fatalf("expected ErrPhotoBadURI, got %v", err)
	}
}

func TestValidatePhoto_CaptionTooLong(t *testing.T) {
	long := strings.Repeat("a", maxPhotoCaption+1)
	if err := ValidatePhoto(validPhotoJSON(t, long)); err != ErrPhotoCaptionLong {
		t.Fatalf("expected ErrPhotoCaptionLong, got %v", err)
	}
}

func TestValidatePhoto_TooBigDecoded(t *testing.T) {
	big := make([]byte, maxPhotoBytes+1)
	enc := base64.StdEncoding.EncodeToString(big)
	if err := ValidatePhoto(`{"v":1,"photo":"data:image/png;base64,` + enc + `"}`); err != ErrPhotoTooBig {
		t.Fatalf("expected ErrPhotoTooBig, got %v", err)
	}
}

func TestValidatePhoto_Success(t *testing.T) {
	if err := ValidatePhoto(validPhotoJSON(t, "a sunny day")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidatePhoto_SuccessNoCaption(t *testing.T) {
	if err := ValidatePhoto(`{"v":1,"photo":"data:image/png;base64,` + base64.StdEncoding.EncodeToString([]byte{0x89, 0x50}) + `"}`); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
