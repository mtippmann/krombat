package main

import (
	"strings"
	"testing"
)

func TestStartupValidatesSecretAndPortBeforeKubernetes(t *testing.T) {
	t.Setenv("SESSION_SECRET", "")
	if err := run(); err == nil || !strings.Contains(err.Error(), "SESSION_SECRET") {
		t.Fatalf("expected explicit missing-secret error, got %v", err)
	}
	t.Setenv("SESSION_SECRET", strings.Repeat("s", 32))
	for _, port := range []string{"not-a-port", "0", "65536"} {
		t.Setenv("PORT", port)
		if err := run(); err == nil || !strings.Contains(err.Error(), "PORT") {
			t.Fatalf("expected explicit port error for %q, got %v", port, err)
		}
	}
}
