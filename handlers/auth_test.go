package handlers

import "testing"

func TestValidPassword(t *testing.T) {
	if validPassword("short") {
		t.Fatal("expected short password to be rejected")
	}
	if !validPassword("long-enough") {
		t.Fatal("expected eight-character password to be accepted")
	}
}
