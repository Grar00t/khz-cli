package board

import (
	"strings"
	"testing"

	"github.com/Grar00t/khz-cli/internal/term"
)

func TestDecideNeverHidesFail(t *testing.T) {
	rows := []Row{
		{State: term.OK, Name: "a"},
		{State: term.FAIL, Name: "b"},
		{State: term.WARN, Name: "c"},
	}
	if d := Decide(rows); d != "FAIL" {
		t.Fatalf("%s", d)
	}
}

func TestDecideReviewAndProceedAndIncomplete(t *testing.T) {
	if Decide([]Row{{State: term.WARN, Name: "s"}}) != "REVIEW" {
		t.Fatal("warn")
	}
	if Decide([]Row{{State: term.OK, Name: "s"}, {State: term.SKIP, Name: "x"}}) != "PROCEED" {
		t.Fatal("ok")
	}
	if Decide([]Row{{State: term.SKIP, Name: "x"}}) != "INCOMPLETE" {
		t.Fatal("skip")
	}
}

func TestRenderAlwaysHasTokensAndSanitizes(t *testing.T) {
	b := Board{
		Headline: "repo@\x1b[41mred\x1b]0;BAD\x07",
		Rows: []Row{
			{State: term.OK, Name: "git", Detail: "clean"},
			{State: term.FAIL, Name: "evil\x1b[41m", Detail: "x"},
		},
		Decision: "FAIL",
		Receipt:  ".khz/receipts/id.json",
	}
	text := Render(b, true, "en", 80)
	if !strings.Contains(text, "OK") || !strings.Contains(text, "FAIL") {
		t.Fatalf("tokens missing:\n%s", text)
	}
	if strings.Contains(text, "[41m") || strings.Contains(text, "]0;BAD") {
		t.Fatalf("untrusted CSI/OSC injected:\n%s", text)
	}
	plain := Render(b, false, "en", 80)
	if strings.Contains(plain, "\x1b") {
		t.Fatalf("plain still colored:\n%s", plain)
	}
}

func TestRenderArabicLabelsKeepStateTokens(t *testing.T) {
	b := Board{
		Headline: "repo@abc",
		Rows:     []Row{{State: term.OK, Name: "git", Detail: "clean"}},
		Decision: "PROCEED",
	}
	text := Render(b, false, "ar", 80)
	if !strings.Contains(text, "OK") {
		t.Fatalf("state token translated: %s", text)
	}
	if !strings.Contains(text, "غيت") && !strings.Contains(text, "القرار") {
		t.Fatalf("expected arabic labels: %s", text)
	}
}
