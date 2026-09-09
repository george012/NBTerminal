package guis

import (
	"strings"
	"testing"
)

func TestNormalizeSessionCustomTitle(t *testing.T) {
	if got, err := normalizeSessionCustomTitle("  Incident 桥接  "); err != nil || got != "Incident 桥接" {
		t.Fatalf("normalized title = %q, err=%v", got, err)
	}
	if got, err := normalizeSessionCustomTitle("   "); err != nil || got != "" {
		t.Fatalf("blank reset = %q, err=%v", got, err)
	}
	if _, err := normalizeSessionCustomTitle("first\nsecond"); err == nil {
		t.Fatal("multiline title was accepted")
	}
	if _, err := normalizeSessionCustomTitle(strings.Repeat("界", sessionCustomTitleMax+1)); err == nil {
		t.Fatal("oversized Unicode title was accepted")
	}
}
