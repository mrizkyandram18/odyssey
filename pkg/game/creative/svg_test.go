package creative

import (
	"strings"
	"testing"
)

func TestValidateSVG(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		wantErr error
	}{
		{
			name:    "valid simple svg",
			payload: `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><circle cx="50" cy="50" r="40" /></svg>`,
			wantErr: nil,
		},
		{
			name:    "empty payload",
			payload: ``,
			wantErr: ErrSVGEmpty,
		},
		{
			name:    "too large",
			payload: `<svg>` + strings.Repeat("a", maxSVGBodyBytes) + `</svg>`,
			wantErr: ErrSVGTooLarge,
		},
		{
			name:    "malformed xml",
			payload: `<svg><circle cx="50" cy="50" r="40" ></svg>`, // missing closing circle tag / improperly closed
			wantErr: ErrSVGMalformed,
		},
		{
			name:    "not svg root",
			payload: `<html><svg></svg></html>`,
			wantErr: ErrSVGRootMissing,
		},
		{
			name:    "multiple roots",
			payload: `<svg></svg><svg></svg>`,
			wantErr: ErrSVGMultipleRoots,
		},
		{
			name:    "script tag",
			payload: `<svg><script>alert(1)</script></svg>`,
			wantErr: ErrSVGDisallowedTag,
		},
		{
			name:    "foreignObject tag",
			payload: `<svg><foreignObject><div>Test</div></foreignObject></svg>`,
			wantErr: ErrSVGDisallowedTag,
		},
		{
			name:    "event handler attr",
			payload: `<svg><circle onload="alert(1)" /></svg>`,
			wantErr: ErrSVGDisallowedAttr,
		},
		{
			name:    "external href",
			payload: `<svg><a href="http://malicious.com"><circle /></a></svg>`,
			wantErr: ErrSVGDisallowedURI,
		},
		{
			name:    "external xlink:href",
			payload: `<svg><image xlink:href="https://malicious.com/img.png" /></svg>`,
			wantErr: ErrSVGDisallowedURI,
		},
		{
			name:    "javascript href",
			payload: `<svg><a href="javascript:alert(1)"><circle /></a></svg>`,
			wantErr: ErrSVGDisallowedURI,
		},
		{
			name:    "data href",
			payload: `<svg><image href="data:image/png;base64,iVBORw0KGgo=" /></svg>`,
			wantErr: ErrSVGDisallowedURI,
		},
		{
			name:    "local fragment href",
			payload: `<svg><use href="#my-shape" /></svg>`,
			wantErr: nil,
		},
		{
			name:    "empty href",
			payload: `<svg><a href=""><circle /></a></svg>`,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSVG(tt.payload)
			if err != tt.wantErr {
				t.Errorf("ValidateSVG() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
