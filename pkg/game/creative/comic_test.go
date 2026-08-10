package creative

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

func TestValidateComic(t *testing.T) {
	validSVG := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><circle cx="5" cy="5" r="4"/></svg>`

	mustJSON := func(v any) string {
		t.Helper()
		b, err := json.Marshal(v)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		return string(b)
	}

	tests := []struct {
		name    string
		payload string
		wantErr error
	}{
		{
			name: "valid two caption panels",
			payload: mustJSON(ComicPayload{
				V: 1,
				Panels: []ComicPanel{
					{Caption: "Once upon a time"},
					{Caption: "The end"},
				},
			}),
			wantErr: nil,
		},
		{
			name: "valid mixed caption and svg",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{
					{Caption: "Hero appears", SVG: validSVG},
					{Caption: "Victory!"},
				},
			}),
			wantErr: nil,
		},
		{
			name: "valid svg-only panels",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{
					{SVG: validSVG},
					{SVG: validSVG},
				},
			}),
			wantErr: nil,
		},
		{
			name:    "empty payload",
			payload: "",
			wantErr: ErrComicEmpty,
		},
		{
			name:    "too large",
			payload: `{"panels":[{"caption":"a"},{"caption":"` + strings.Repeat("x", maxComicBodyBytes) + `"}]}`,
			wantErr: ErrComicTooLarge,
		},
		{
			name:    "malformed json",
			payload: `{not json`,
			wantErr: ErrComicMalformed,
		},
		{
			name: "one panel",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{{Caption: "alone"}},
			}),
			wantErr: ErrComicPanelCount,
		},
		{
			name: "five panels",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{
					{Caption: "1"}, {Caption: "2"}, {Caption: "3"},
					{Caption: "4"}, {Caption: "5"},
				},
			}),
			wantErr: ErrComicPanelCount,
		},
		{
			name: "empty panel",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{
					{Caption: "ok"},
					{Caption: "", SVG: ""},
				},
			}),
			wantErr: ErrComicPanelEmpty,
		},
		{
			name: "caption too long",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{
					{Caption: strings.Repeat("a", maxComicCaption+1)},
					{Caption: "b"},
				},
			}),
			wantErr: ErrComicCaptionTooLong,
		},
		{
			name: "disallowed svg script",
			payload: mustJSON(ComicPayload{
				Panels: []ComicPanel{
					{Caption: "a", SVG: `<svg><script>alert(1)</script></svg>`},
					{Caption: "b"},
				},
			}),
			wantErr: ErrSVGDisallowedTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateComic(tt.payload)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("expected nil, got %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}
