package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
)

// GenerateOTP returns a 6-digit numeric code as a string.
func GenerateOTP() string {
	n := mrand.Intn(900000) + 100000
	return fmt.Sprintf("%d", n)
}

// HashOTP hashes an OTP so the plaintext is never stored.
func HashOTP(otp string) string {
	sum := sha256.Sum256([]byte(otp))
	return hex.EncodeToString(sum[:])
}

// GenerateSessionToken returns a random 64-char hex token for cookie-based sessions.
func GenerateSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
