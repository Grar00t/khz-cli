package redact

import "strings"

const replacement = "<redacted>"

var sensitive = map[string]struct{}{
	"password":      {},
	"passwd":        {},
	"token":         {},
	"api-key":       {},
	"apikey":        {},
	"secret":        {},
	"authorization": {},
}

func normalizeKey(s string) string {
	s = strings.TrimLeft(s, "-")
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func isSensitive(s string) bool {
	_, ok := sensitive[normalizeKey(s)]
	return ok
}

// Args performs best-effort redaction without modifying the caller's slice.
func Args(args []string) []string {
	if len(args) == 0 {
		return []string{}
	}
	out := append([]string(nil), args...)
	for i := 0; i < len(out); i++ {
		arg := out[i]
		if key, _, ok := strings.Cut(arg, "="); ok && isSensitive(key) {
			out[i] = key + "=" + replacement
			continue
		}
		if isSensitive(arg) && i+1 < len(out) {
			out[i+1] = replacement
			i++
		}
	}
	return out
}
