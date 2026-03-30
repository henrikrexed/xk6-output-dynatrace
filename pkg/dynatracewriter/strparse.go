package dynatracewriter

import (
	"strconv"
	"strings"
)

// parseKeyValuePairs parses a comma-separated key=value string into a map.
// Supports nested keys with dots (e.g., "headers.X-Key=value").
// Boolean and numeric values are converted to their native types.
func parseKeyValuePairs(s string) map[string]interface{} {
	result := make(map[string]interface{})
	if s == "" {
		return result
	}

	for _, pair := range strings.Split(s, ",") {
		pair = strings.TrimSpace(pair)
		idx := strings.Index(pair, "=")
		if idx < 0 {
			continue
		}
		key := strings.TrimSpace(pair[:idx])
		val := strings.TrimSpace(pair[idx+1:])

		// Handle nested keys like "headers.X-Header"
		if parts := strings.SplitN(key, ".", 2); len(parts) == 2 {
			parent := parts[0]
			child := parts[1]
			m, ok := result[parent].(map[string]interface{})
			if !ok {
				m = make(map[string]interface{})
				result[parent] = m
			}
			m[child] = val
			continue
		}

		// Try bool
		if b, err := strconv.ParseBool(val); err == nil {
			result[key] = b
			continue
		}

		// Store as string
		result[key] = val
	}

	return result
}
