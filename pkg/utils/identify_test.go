package utils

import (
	"context"
	"testing"
)

func TestTLSIssuer(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	expected := GetTLSIssuer(ctx)
	if expected != "" {
		t.Fatalf("expected empty, got %s", expected)
	}

	expected = "bar"
	ctx = SetTLSIssuer(ctx, expected)

	got := GetTLSIssuer(ctx)
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}

func TestTLSCommonName(t *testing.T) {
	ctx := context.Background()
	defer ctx.Done()

	expected := GetTLSCommonName(ctx)
	if expected != "" {
		t.Fatalf("expected empty, got %s", expected)
	}

	expected = "bar"
	ctx = SetTLSCommonName(ctx, expected)

	got := GetTLSCommonName(ctx)
	if got != expected {
		t.Fatalf("expected %s, got %s", expected, got)
	}
}
