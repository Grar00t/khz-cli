package redact

import (
	"reflect"
	"testing"
)

func TestArgsRedactsFlagValue(t *testing.T) {
	in := []string{"tool", "--password", "s3cret", "--file", "x"}
	got := Args(in)
	want := []string{"tool", "--password", "***", "--file", "x"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v", got)
	}
	if in[2] != "s3cret" {
		t.Fatal("must not mutate input")
	}
}

func TestArgsRedactsEqualsForm(t *testing.T) {
	got := Args([]string{"--token=abcd", "--ok=1"})
	if got[0] != "--token=***" || got[1] != "--ok=1" {
		t.Fatalf("got %#v", got)
	}
}

func TestArgsRedactsAllListedTokens(t *testing.T) {
	keys := []string{
		"--passwd", "--token", "--api-key", "--apikey", "--secret", "--authorization", "--password",
	}
	for _, k := range keys {
		got := Args([]string{k, "VALUE"})
		if got[1] != "***" {
			t.Fatalf("%s: %#v", k, got)
		}
	}
}

func TestArgsLeavesOrdinaryFlags(t *testing.T) {
	in := []string{"git", "status", "--porcelain=v2"}
	got := Args(in)
	if !reflect.DeepEqual(got, in) {
		t.Fatalf("got %#v", got)
	}
}
