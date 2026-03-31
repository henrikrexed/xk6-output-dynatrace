package dynatracewriter

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeURL_NumericIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"/v2/pet/123", "/v2/pet/{id}"},
		{"/v2/pet/9223372036854284000", "/v2/pet/{id}"},
		{"/api/v1/users/42/orders/99", "/api/v1/users/{id}/orders/{id}"},
		{"/pet/0", "/pet/{id}"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeURL(tt.input), "input: %s", tt.input)
	}
}

func TestNormalizeURL_UUIDs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"/users/550e8400-e29b-41d4-a716-446655440000", "/users/{id}"},
		{"/api/550e8400-e29b-41d4-a716-446655440000/detail", "/api/{id}/detail"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeURL(tt.input), "input: %s", tt.input)
	}
}

func TestNormalizeURL_LongHex(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		{"/commit/8249de92ab", "/commit/{id}"},
		{"/blob/deadbeef01234567", "/blob/{id}"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeURL(tt.input), "input: %s", tt.input)
	}
}

func TestNormalizeURL_MixedPaths(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected string
	}{
		// Short hex that should NOT be replaced (<=8 chars)
		{"/api/v2/abcdef12/info", "/api/v2/abcdef12/info"},
		// Non-dynamic segments preserved
		{"/api/v2/pet/findByStatus", "/api/v2/pet/findByStatus"},
		// Empty path
		{"", ""},
		// Root
		{"/", "/"},
		// Query string preserved
		{"/pet/123?format=json", "/pet/{id}?format=json"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.expected, normalizeURL(tt.input), "input: %s", tt.input)
	}
}

func TestNormalizeURL_FullURLs(t *testing.T) {
	t.Parallel()
	// Full URLs with scheme and host (k6 sometimes stores full URLs)
	input := "https://petstore.example.com/v2/pet/9223372036854284000"
	result := normalizeURL(input)
	assert.Equal(t, "https://petstore.example.com/v2/pet/{id}", result)
}

func TestTruncateDimensionValue(t *testing.T) {
	t.Parallel()

	short := "/api/v1/test"
	assert.Equal(t, short, truncateDimensionValue(short))

	long := strings.Repeat("a", 300)
	result := truncateDimensionValue(long)
	assert.Equal(t, 250, len(result))
}

func TestNormalizeURL_TruncatesLongResult(t *testing.T) {
	t.Parallel()
	// Build a URL that, even after normalization, exceeds 250 chars
	long := "/" + strings.Repeat("segment/", 40) + "end"
	result := normalizeURL(long)
	assert.LessOrEqual(t, len(result), maxDimensionValueLen)
}

func TestNormalizeURL_DisabledByEnv(t *testing.T) {
	t.Setenv("K6_DYNATRACE_URL_NORMALIZATION", "false")

	input := "/v2/pet/123"
	// When disabled, numeric IDs should NOT be replaced
	result := normalizeURL(input)
	assert.Equal(t, "/v2/pet/123", result)
}
