package services

import (
	"strings"
	"testing"
)

func TestBuildDownloadSearchFilter(t *testing.T) {
	sql, args := buildDownloadSearchFilter("  silo s03e03  ")
	if len(args) != 4 {
		t.Fatalf("expected 4 args for 2 terms, got %d (%v)", len(args), args)
	}
	if args[0] != "%silo%" || args[2] != "%s03e03%" {
		t.Fatalf("unexpected patterns: %v", args)
	}
	if !strings.Contains(sql, " AND ") {
		t.Fatalf("expected terms to be AND-ed, got %q", sql)
	}

	_, args = buildDownloadSearchFilter("Mr.Inbetween (2018)")
	if len(args) != 6 {
		t.Fatalf("expected separators to split into 3 terms, got %v", args)
	}
	if args[0] != "%mr%" || args[2] != "%inbetween%" || args[4] != "%2018%" {
		t.Fatalf("unexpected patterns: %v", args)
	}

	if sql, _ := buildDownloadSearchFilter("   "); sql != "" {
		t.Fatalf("expected empty filter for blank query, got %q", sql)
	}
}
