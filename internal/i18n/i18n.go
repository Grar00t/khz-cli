package i18n

// T returns a human-visible string. JSON keys, hashes, paths, and state tokens
// are never translated.
func T(lang, key string) string {
	if lang == "ar" {
		if s, ok := ar[key]; ok {
			return s
		}
	}
	if s, ok := en[key]; ok {
		return s
	}
	return key
}

var en = map[string]string{
	"git":        "git",
	"source":     "source",
	"checks":     "checks",
	"policy":     "policy",
	"doctor":     "doctor",
	"run":        "run",
	"decision":   "DECISION",
	"receipt":    "RECEIPT",
	"clean":      "clean",
	"dirty":      "dirty",
	"unborn":     "unborn branch",
	"detached":   "detached HEAD",
	"not_repo":   "not a git repository",
	"none":       "none recorded",
	"proceed":    "PROCEED",
	"review":     "REVIEW",
	"fail":       "FAIL",
	"incomplete": "INCOMPLETE",
	"protected":  "protected branch (local policy)",
	"ok":         "ok",
}

var ar = map[string]string{
	"git":        "غيت",
	"source":     "المصدر",
	"checks":     "الفحوصات",
	"policy":     "السياسة",
	"doctor":     "التشخيص",
	"run":        "تشغيل",
	"decision":   "القرار",
	"receipt":    "الإيصال",
	"clean":      "نظيف",
	"dirty":      "غير نظيف",
	"unborn":     "فرع بلا التزامات",
	"detached":   "رأس منفصل",
	"not_repo":   "ليس مستودع غيت",
	"none":       "لا سجلات",
	"proceed":    "PROCEED",
	"review":     "REVIEW",
	"fail":       "FAIL",
	"incomplete": "INCOMPLETE",
	"protected":  "فرع محمي (سياسة محلية)",
	"ok":         "سليم",
}
