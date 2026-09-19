package redact

import (
	"reflect"
	"testing"
)

func TestArgs(t *testing.T) {
	in := []string{"login", "--password", "alpha", "--api-key=beta", "token=gamma", "safe"}
	want := []string{"login", "--password", "<redacted>", "--api-key=<redacted>", "token=<redacted>", "safe"}
	got := Args(in)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v want %#v", got, want)
	}
	if in[2] != "alpha" {
		t.Fatal("input slice mutated")
	}
}

func TestAuthorizationCaseAndUnderscore(t *testing.T) {
	got := Args([]string{"--AUTHORIZATION", "Bearer x", "--api_key", "abc"})
	if got[1] != "<redacted>" || got[3] != "<redacted>" {
		t.Fatalf("redaction failed: %#v", got)
	}
}
