package redact

import (
	"strings"
)

// Tokens that, when present in a flag or key, cause the associated value to be redacted.
// Matching is case-insensitive and ignores leading dashes.
var sensitive = []string{
	"password",
	"passwd",
	"token",
	"api-key",
	"apikey",
	"secret",
	"authorization",
}

const redacted = "***"

func isSensitiveKey(key string) bool {
	k := strings.ToLower(strings.TrimLeft(key, "-"))
	k = strings.ReplaceAll(k, "_", "-")
	for _, s := range sensitive {
		if k == s || strings.Contains(k, s) {
			return true
		}
	}
	return false
}

// Args returns a copy of argv with best-effort secret redaction.
// It cannot guarantee that arbitrary user-provided secrets will never appear.
func Args(argv []string) []string {
	out := make([]string, len(argv))
	copy(out, argv)
	for i := 0; i < len(out); i++ {
		a := out[i]
		if eq := strings.IndexByte(a, '='); eq > 0 && strings.HasPrefix(a, "-") {
			key, val := a[:eq], a[eq+1:]
			if isSensitiveKey(key) && val != "" {
				out[i] = key + "=" + redacted
			}
			continue
		}
		if isSensitiveKey(a) && i+1 < len(out) && !strings.HasPrefix(out[i+1], "-") {
			out[i+1] = redacted
			i++
		}
	}
	return out
}
