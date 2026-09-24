package handlers

import (
	"net/mail"
	"testing"
)

func TestNormalizeShopProfile(t *testing.T) {
	shopName, ownerName, plan := normalizeShopProfile("  My Shop  ", "  Rahul Verma  ", " Pro ")
	if shopName != "My Shop" {
		t.Fatalf("expected normalized shop name, got %q", shopName)
	}
	if ownerName != "Rahul Verma" {
		t.Fatalf("expected normalized owner name, got %q", ownerName)
	}
	if plan != "pro" {
		t.Fatalf("expected normalized plan, got %q", plan)
	}
}

func TestNormalizeShopProfileDefaults(t *testing.T) {
	shopName, ownerName, plan := normalizeShopProfile(" ", " ", "unknown")
	if shopName != "My Shop" {
		t.Fatalf("expected default shop name, got %q", shopName)
	}
	if ownerName != "" {
		t.Fatalf("expected empty owner name, got %q", ownerName)
	}
	if plan != "trial" {
		t.Fatalf("expected default plan, got %q", plan)
	}
}

func TestRecoveryEmailValidation(t *testing.T) {
	if _, err := mail.ParseAddress("owner@example.com"); err != nil {
		t.Fatalf("expected valid email, got %v", err)
	}
	if _, err := mail.ParseAddress("not-an-email"); err == nil {
		t.Fatal("expected invalid email to be rejected")
	}
}
