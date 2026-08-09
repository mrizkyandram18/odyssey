package creative

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
)

var (
	ErrSVGEmpty         = errors.New("svg payload is empty")
	ErrSVGTooLarge      = errors.New("svg payload exceeds maximum size")
	ErrSVGMalformed     = errors.New("svg is malformed xml")
	ErrSVGRootMissing   = errors.New("svg root element must be <svg>")
	ErrSVGMultipleRoots = errors.New("svg has multiple root elements")
	ErrSVGDisallowedTag = errors.New("svg contains disallowed element")
	ErrSVGDisallowedAttr= errors.New("svg contains disallowed attribute")
	ErrSVGDisallowedURI = errors.New("svg contains disallowed external reference")
)

const maxSVGBodyBytes = 250 * 1024 // 250 KiB

// ValidateSVG validates that a given raw SVG payload is safe to store and render.
// It applies a strict allowlist/denylist:
// - Max 250 KiB
// - Must be well-formed XML with exactly one root <svg> element.
// - Must not contain <script> or <foreignObject>.
// - Must not contain any attribute starting with 'on' (event handlers).
// - Any 'href' or 'xlink:href' must be local (empty or start with '#').
func ValidateSVG(payload string) error {
	if payload == "" {
		return ErrSVGEmpty
	}
	if len(payload) > maxSVGBodyBytes {
		return ErrSVGTooLarge
	}

	decoder := xml.NewDecoder(strings.NewReader(payload))
	// Keep Strict = true to enforce well-formed XML

	rootCount := 0
	depth := 0

	for {
		t, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ErrSVGMalformed
		}

		switch el := t.(type) {
		case xml.StartElement:
			if depth == 0 {
				rootCount++
				if rootCount > 1 {
					return ErrSVGMultipleRoots
				}
				if strings.ToLower(el.Name.Local) != "svg" {
					return ErrSVGRootMissing
				}
			}
			depth++

			tagName := strings.ToLower(el.Name.Local)
			if tagName == "script" || tagName == "foreignobject" {
				return ErrSVGDisallowedTag
			}

			for _, attr := range el.Attr {
				attrName := strings.ToLower(attr.Name.Local)
				if strings.HasPrefix(attrName, "on") {
					return ErrSVGDisallowedAttr
				}

				if attrName == "href" || attrName == "xlink:href" || (attr.Name.Space == "xlink" && attrName == "href") {
					val := strings.TrimSpace(attr.Value)
					if val != "" && !strings.HasPrefix(val, "#") {
						return ErrSVGDisallowedURI
					}
				}
			}

		case xml.EndElement:
			depth--
		}
	}

	if rootCount == 0 {
		return ErrSVGRootMissing
	}

	if depth != 0 {
		return ErrSVGMalformed
	}

	return nil
}
