package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateSessionToken returns a random 64-char hex token for cookie-based sessions.
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
