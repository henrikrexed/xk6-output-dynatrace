package dynatracewriter

import (
	"os"
	"regexp"
	"strings"
)

const maxDimensionValueLen = 250

var (
	// Matches pure numeric path segments like /123, /9223372036854284000
	numericSegmentRe = regexp.MustCompile(`^[0-9]+$`)
	// Matches UUIDs like 550e8400-e29b-41d4-a716-446655440000
	uuidSegmentRe = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	// Matches long hex strings (>8 chars) like 8249de92ab, deadbeef01234567
	longHexSegmentRe = regexp.MustCompile(`^[0-9a-fA-F]{9,}$`)
)

// normalizeURL replaces dynamic path segments with {id} placeholders to reduce
// metric cardinality. Pure numeric IDs, UUIDs, and long hex strings are all
// replaced. This prevents MINT protocol line length overflow and Dynatrace
// dimension value length rejections.
func normalizeURL(rawURL string) string {
	if !isURLNormalizationEnabled() {
		return truncateDimensionValue(rawURL)
	}

	// Handle query strings: normalize only the path portion
	path := rawURL
	query := ""
	if idx := strings.IndexByte(rawURL, '?'); idx >= 0 {
		path = rawURL[:idx]
		query = rawURL[idx:]
	}

	segments := strings.Split(path, "/")
	for i, seg := range segments {
		if seg == "" {
			continue
		}
		if numericSegmentRe.MatchString(seg) ||
			uuidSegmentRe.MatchString(seg) ||
			longHexSegmentRe.MatchString(seg) {
			segments[i] = "{id}"
		}
	}

	return truncateDimensionValue(strings.Join(segments, "/") + query)
}

// truncateDimensionValue ensures a dimension value does not exceed the
// Dynatrace 250-character limit.
func truncateDimensionValue(v string) string {
	if len(v) > maxDimensionValueLen {
		return v[:maxDimensionValueLen]
	}
	return v
}

// isURLNormalizationEnabled checks the K6_DYNATRACE_URL_NORMALIZATION env var.
// Defaults to true if unset or empty.
func isURLNormalizationEnabled() bool {
	v := os.Getenv("K6_DYNATRACE_URL_NORMALIZATION")
	return v == "" || v == "true" || v == "1"
}
